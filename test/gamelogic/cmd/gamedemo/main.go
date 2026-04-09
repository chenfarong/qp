// gamedemo：游戏逻辑命令行演示。默认读取仓库根目录下 configs/config.yaml。
//
// 用法：
//
//	go run ./test/gamelogic/cmd/gamedemo
//	go run ./test/gamelogic/cmd/gamedemo -username u -password p -character 1
//	go run ./test/gamelogic/cmd/gamedemo -username u -password p -character 507f1f77bcf86cd799439011
//
// 三者皆不提供：自动注册账号并创建一名角色，凭据与角色 ID 会打印在终端。
// 三者均提供：先登录，再按 -character 解析角色（1-based 序号，或 24 位十六进制角色 ObjectID）。
package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aoyo/qp/internal/gamelogic"
	"github.com/aoyo/qp/internal/gamelogic/actor"
	authsvc "github.com/aoyo/qp/internal/ssoauth/service"
	"github.com/aoyo/qp/pkg/db"
	"gopkg.in/yaml.v3"
)

type configYAML struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"database"`
	Jwt struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
}

func loadConfig(path string) (*configYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c configYAML
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func mongoURI(c *configYAML) string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Dbname,
	)
}

func randomPassword(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("DemoPwd%d", time.Now().UnixNano())
	}
	const alphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKMNPQRSTUVWXYZ23456789"
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

func isHexObjectID(s string) bool {
	if len(s) != 24 {
		return false
	}
	for _, r := range s {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return false
		}
	}
	return true
}

func resolveCharacterID(userIDHex, spec string, charSvc *actor.CharacterService) (string, error) {
	if isHexObjectID(spec) {
		ch, err := charSvc.GetCharacterByID(spec)
		if err != nil {
			return "", fmt.Errorf("角色不存在或无效: %w", err)
		}
		if ch.UserID.Hex() != userIDHex {
			return "", fmt.Errorf("该角色不属于当前用户")
		}
		return ch.ID.Hex(), nil
	}
	idx, err := strconv.Atoi(strings.TrimSpace(spec))
	if err != nil || idx < 1 {
		return "", fmt.Errorf("-character 须为 1 起的序号，或 24 位十六进制角色 ID")
	}
	chars, err := charSvc.GetCharactersByUserID(userIDHex)
	if err != nil {
		return "", err
	}
	if idx > len(chars) {
		return "", fmt.Errorf("序号 %d 超出该用户角色数量 (%d)", idx, len(chars))
	}
	return chars[idx-1].ID.Hex(), nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[gamedemo] ")

	configPath := flag.String("config", "", "配置文件路径（默认：从当前目录向上查找 configs/config.yaml）")
	username := flag.String("username", "", "用户名")
	password := flag.String("password", "", "密码")
	character := flag.String("character", "", "角色：1-based 序号，或 24 位 hex 角色 ID")
	flag.Parse()

	cfgFile := *configPath
	if cfgFile == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("获取工作目录失败: %v", err)
		}
		for i := 0; i < 8; i++ {
			candidate := filepath.Join(dir, "configs", "config.yaml")
			if _, err := os.Stat(candidate); err == nil {
				cfgFile = candidate
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if cfgFile == "" {
			log.Fatal("未找到 configs/config.yaml，请用 -config 指定路径")
		}
	}

	cfg, err := loadConfig(cfgFile)
	if err != nil {
		log.Fatalf("读取配置失败 (%s): %v", cfgFile, err)
	}
	log.Printf("使用配置: %s", cfgFile)

	dbInst, err := db.InitDB(mongoURI(cfg))
	if err != nil {
		log.Fatalf("连接 MongoDB 失败: %v", err)
	}
	defer dbInst.Close()

	auth := authsvc.NewAuthService(dbInst, cfg.Jwt.Secret, cfg.Jwt.ExpireHours, cfg.Database.Dbname)
	game := gamelogic.NewApp(dbInst, cfg.Database.Dbname)

	auto := *username == "" && *password == "" && *character == ""
	explicit := *username != "" && *password != "" && *character != ""

	if !auto && !explicit {
		log.Fatal("请「同时」提供 -username、-password、-character，或「三者都不提供」以进入自动注册演示模式")
	}

	var userIDHex string
	var plainPassword string

	if auto {
		rb := make([]byte, 4)
		if _, err := rand.Read(rb); err != nil {
			log.Printf("随机数读取失败，使用纯时间戳作为后缀: %v", err)
		}
		suffix := hex.EncodeToString(rb)
		u := fmt.Sprintf("qp_auto_%d_%s", time.Now().UnixNano(), suffix)
		plainPassword = randomPassword(12)
		email := fmt.Sprintf("%s@qp-demo.local", u)

		log.Printf("未提供账号参数，开始自动注册: username=%s", u)
		regResp, err := auth.Register(authsvc.RegisterRequest{
			Username: u,
			Password: plainPassword,
			Email:    email,
			Nickname: u,
		})
		if err != nil {
			log.Fatalf("自动注册失败: %v", err)
		}
		userIDHex = regResp.UserInfo.ID.Hex()
		log.Printf("注册成功: user_id=%s, token 已签发（省略打印）", userIDHex)

		charName := fmt.Sprintf("角色_%s", time.Now().Format("150405"))
		log.Printf("自动创建角色: name=%s", charName)
		createResp, err := game.CharacterService.CreateCharacter(actor.CreateCharacterRequest{
			UserID: userIDHex,
			Name:   charName,
		})
		if err != nil {
			log.Fatalf("创建角色失败: %v", err)
		}
		charID := createResp.Character.ID.Hex()
		log.Printf("创建角色成功: character_id=%s, name=%s, level=%d",
			charID, createResp.Character.Name, createResp.Character.Level)

		log.Println("---------- 自动演示凭据（请自行保存） ----------")
		log.Printf("用户名: %s", u)
		log.Printf("密码: %s", plainPassword)
		log.Printf("角色 ID: %s", charID)
		log.Println("下次可执行: go run ./test/gamelogic/cmd/gamedemo -username <上> -password <上> -character 1")
		log.Println("-----------------------------------------------")
		return
	}

	log.Printf("使用已有账号登录: username=%s", *username)
	loginResp, err := auth.Login(authsvc.LoginRequest{
		Username: *username,
		Password: *password,
	})
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	userIDHex = loginResp.UserInfo.ID.Hex()
	log.Printf("登录成功: user_id=%s", userIDHex)

	charID, err := resolveCharacterID(userIDHex, *character, game.CharacterService)
	if err != nil {
		log.Fatalf("解析角色失败: %v", err)
	}
	ch, err := game.CharacterService.GetCharacterByID(charID)
	if err != nil {
		log.Fatalf("加载角色失败: %v", err)
	}
	log.Printf("选中角色: character_id=%s, name=%s, level=%d, items=%d",
		charID, ch.Name, ch.Level, len(ch.Items))
}
