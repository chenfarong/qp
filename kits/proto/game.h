#ifndef GAME_H
#define GAME_H

#include <cstdint>
#include <string>
#include <vector>
#include <memory>
#include <stdexcept>

namespace game {

// Message types
enum class MessageType {
    MSG_TYPE_UNKNOWN = 0,
    MSG_TYPE_AUTH_REGISTER = 1,
    MSG_TYPE_AUTH_LOGIN = 2,
    MSG_TYPE_AUTH_VALIDATE = 3,
    MSG_TYPE_GAME_CREATE_CHARACTER = 4,
    MSG_TYPE_GAME_GET_CHARACTERS = 5,
    MSG_TYPE_GAME_GET_CHARACTER = 6,
    MSG_TYPE_GAME_UPDATE_CHARACTER_STATUS = 7,
    MSG_TYPE_GAME_BATTLE = 8,
    MSG_TYPE_BILL_GET_TOKEN_BALANCE = 9,
    MSG_TYPE_BILL_ADD_TOKEN = 10,
    MSG_TYPE_BILL_REMOVE_TOKEN = 11,
    MSG_TYPE_BILL_CREATE_PAYMENT = 12,
    MSG_TYPE_BILL_GET_PAYMENT = 13,
    MSG_TYPE_RESPONSE = 100
};

// Message message
struct Message {
    MessageType type;

    // Default constructor
    Message() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize type
        buffer.push_back(8); // Field key: 8 (field number 1, wire type 0)
            // Enum encoding for type
            uint64_t value = static_cast<uint64_t>(type);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<Message> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<Message>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // type
                    // Enum decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->type = static_cast<MessageType>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// AuthRegisterRequest message
struct AuthRegisterRequest {
    std::string username;
    std::string password;
    std::string email;
    std::string nickname;

    // Default constructor
    AuthRegisterRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize username
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = username.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), username.begin(), username.end());

        // Serialize password
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = password.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), password.begin(), password.end());

        // Serialize email
        buffer.push_back(26); // Field key: 26 (field number 3, wire type 2)
            // String encoding
            uint64_t length = email.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), email.begin(), email.end());

        // Serialize nickname
        buffer.push_back(34); // Field key: 34 (field number 4, wire type 2)
            // String encoding
            uint64_t length = nickname.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), nickname.begin(), nickname.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<AuthRegisterRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<AuthRegisterRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // username
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->username = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // password
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->password = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // email
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->email = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 4: { // nickname
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->nickname = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// AuthLoginRequest message
struct AuthLoginRequest {
    std::string username;
    std::string password;

    // Default constructor
    AuthLoginRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize username
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = username.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), username.begin(), username.end());

        // Serialize password
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = password.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), password.begin(), password.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<AuthLoginRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<AuthLoginRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // username
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->username = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // password
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->password = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// AuthValidateRequest message
struct AuthValidateRequest {
    std::string token;

    // Default constructor
    AuthValidateRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize token
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = token.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token.begin(), token.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<AuthValidateRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<AuthValidateRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // token
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameCreateCharacterRequest message
struct GameCreateCharacterRequest {
    std::string user_id;
    std::string name;

    // Default constructor
    GameCreateCharacterRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize name
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = name.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), name.begin(), name.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameCreateCharacterRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameCreateCharacterRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // name
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->name = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameGetCharactersRequest message
struct GameGetCharactersRequest {
    std::string user_id;

    // Default constructor
    GameGetCharactersRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameGetCharactersRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameGetCharactersRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameGetCharacterRequest message
struct GameGetCharacterRequest {
    std::string id;

    // Default constructor
    GameGetCharacterRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), id.begin(), id.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameGetCharacterRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameGetCharacterRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameUpdateCharacterStatusRequest message
struct GameUpdateCharacterStatusRequest {
    std::string id;
    int32_t status;

    // Default constructor
    GameUpdateCharacterStatusRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), id.begin(), id.end());

        // Serialize status
        buffer.push_back(16); // Field key: 16 (field number 2, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(status);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameUpdateCharacterStatusRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameUpdateCharacterStatusRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // status
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->status = static_cast<int32_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameBattleRequest message
struct GameBattleRequest {
    std::string character_id;
    int32_t enemy_level;

    // Default constructor
    GameBattleRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize character_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = character_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), character_id.begin(), character_id.end());

        // Serialize enemy_level
        buffer.push_back(16); // Field key: 16 (field number 2, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(enemy_level);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameBattleRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameBattleRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // character_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->character_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // enemy_level
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->enemy_level = static_cast<int32_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillGetTokenBalanceRequest message
struct BillGetTokenBalanceRequest {
    std::string user_id;
    std::string token_type;

    // Default constructor
    BillGetTokenBalanceRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize token_type
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillGetTokenBalanceRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillGetTokenBalanceRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillAddTokenRequest message
struct BillAddTokenRequest {
    std::string user_id;
    std::string token_type;
    int64_t amount;

    // Default constructor
    BillAddTokenRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize token_type
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        // Serialize amount
        buffer.push_back(24); // Field key: 24 (field number 3, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillAddTokenRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillAddTokenRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->amount = static_cast<int64_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillRemoveTokenRequest message
struct BillRemoveTokenRequest {
    std::string user_id;
    std::string token_type;
    int64_t amount;

    // Default constructor
    BillRemoveTokenRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize token_type
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        // Serialize amount
        buffer.push_back(24); // Field key: 24 (field number 3, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillRemoveTokenRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillRemoveTokenRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->amount = static_cast<int64_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillCreatePaymentRequest message
struct BillCreatePaymentRequest {
    std::string user_id;
    int64_t amount;
    std::string currency;
    std::string token_type;
    int64_t token_amount;
    std::string payment_method;

    // Default constructor
    BillCreatePaymentRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize amount
        buffer.push_back(16); // Field key: 16 (field number 2, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize currency
        buffer.push_back(26); // Field key: 26 (field number 3, wire type 2)
            // String encoding
            uint64_t length = currency.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), currency.begin(), currency.end());

        // Serialize token_type
        buffer.push_back(34); // Field key: 34 (field number 4, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        // Serialize token_amount
        buffer.push_back(40); // Field key: 40 (field number 5, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(token_amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize payment_method
        buffer.push_back(50); // Field key: 50 (field number 6, wire type 2)
            // String encoding
            uint64_t length = payment_method.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), payment_method.begin(), payment_method.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillCreatePaymentRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillCreatePaymentRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->amount = static_cast<int64_t>(value);
                    break;
                }
                case 3: { // currency
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->currency = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 4: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 5: { // token_amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->token_amount = static_cast<int64_t>(value);
                    break;
                }
                case 6: { // payment_method
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->payment_method = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillGetPaymentRequest message
struct BillGetPaymentRequest {
    std::string order_id;

    // Default constructor
    BillGetPaymentRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize order_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = order_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), order_id.begin(), order_id.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillGetPaymentRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillGetPaymentRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // order_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->order_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// Response message
struct Response {
    int32_t code;
    std::string message;

    // Default constructor
    Response() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize code
        buffer.push_back(8); // Field key: 8 (field number 1, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(code);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize message
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = message.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), message.begin(), message.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<Response> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<Response>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // code
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->code = static_cast<int32_t>(value);
                    break;
                }
                case 2: { // message
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->message = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// AuthResponse message
struct AuthResponse {
    std::string token;
    std::string user_id;
    std::string username;
    std::string nickname;

    // Default constructor
    AuthResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize token
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = token.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token.begin(), token.end());

        // Serialize user_id
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize username
        buffer.push_back(26); // Field key: 26 (field number 3, wire type 2)
            // String encoding
            uint64_t length = username.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), username.begin(), username.end());

        // Serialize nickname
        buffer.push_back(34); // Field key: 34 (field number 4, wire type 2)
            // String encoding
            uint64_t length = nickname.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), nickname.begin(), nickname.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<AuthResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<AuthResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // token
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // username
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->username = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 4: { // nickname
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->nickname = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameCharacterResponse message
struct GameCharacterResponse {
    std::string id;
    std::string user_id;
    std::string name;
    int32_t level;
    int32_t exp;
    int32_t hp;
    int32_t mp;
    int32_t attack;
    int32_t defense;
    int32_t status;

    // Default constructor
    GameCharacterResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), id.begin(), id.end());

        // Serialize user_id
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize name
        buffer.push_back(26); // Field key: 26 (field number 3, wire type 2)
            // String encoding
            uint64_t length = name.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), name.begin(), name.end());

        // Serialize level
        buffer.push_back(32); // Field key: 32 (field number 4, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(level);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize exp
        buffer.push_back(40); // Field key: 40 (field number 5, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(exp);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize hp
        buffer.push_back(48); // Field key: 48 (field number 6, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(hp);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize mp
        buffer.push_back(56); // Field key: 56 (field number 7, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(mp);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize attack
        buffer.push_back(64); // Field key: 64 (field number 8, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(attack);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize defense
        buffer.push_back(72); // Field key: 72 (field number 9, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(defense);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize status
        buffer.push_back(80); // Field key: 80 (field number 10, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(status);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameCharacterResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameCharacterResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // name
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->name = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 4: { // level
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->level = static_cast<int32_t>(value);
                    break;
                }
                case 5: { // exp
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->exp = static_cast<int32_t>(value);
                    break;
                }
                case 6: { // hp
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->hp = static_cast<int32_t>(value);
                    break;
                }
                case 7: { // mp
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->mp = static_cast<int32_t>(value);
                    break;
                }
                case 8: { // attack
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->attack = static_cast<int32_t>(value);
                    break;
                }
                case 9: { // defense
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->defense = static_cast<int32_t>(value);
                    break;
                }
                case 10: { // status
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->status = static_cast<int32_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameCharactersResponse message
struct GameCharactersResponse {
    std::vector<GameCharacterResponse> characters;

    // Default constructor
    GameCharactersResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize characters (repeated)
        for (const auto& item : characters) {
            // Field key: 10 (field number 1, wire type 2)
            buffer.push_back(10);
            // Nested message encoding for GameCharacterResponse
            std::vector<uint8_t> nestedData = item.Serialize();
            uint64_t length = nestedData.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), nestedData.begin(), nestedData.end());
        }

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameCharactersResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameCharactersResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // characters
                    // Nested message decoding for GameCharacterResponse
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        std::vector<uint8_t> nestedData(buffer.begin() + offset, buffer.begin() + offset + length);
                        msg->characters.push_back(*GameCharacterResponse::Deserialize(nestedData));
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// GameBattleResponse message
struct GameBattleResponse {
    std::string character_id;
    int32_t enemy_level;
    bool victory;
    int32_t exp_gained;
    int32_t gold_gained;

    // Default constructor
    GameBattleResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize character_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = character_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), character_id.begin(), character_id.end());

        // Serialize enemy_level
        buffer.push_back(16); // Field key: 16 (field number 2, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(enemy_level);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize victory
        buffer.push_back(24); // Field key: 24 (field number 3, wire type 0)
            // Varint encoding for bool
            uint64_t value = static_cast<uint64_t>(victory);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize exp_gained
        buffer.push_back(32); // Field key: 32 (field number 4, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(exp_gained);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize gold_gained
        buffer.push_back(40); // Field key: 40 (field number 5, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(gold_gained);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<GameBattleResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<GameBattleResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // character_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->character_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // enemy_level
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->enemy_level = static_cast<int32_t>(value);
                    break;
                }
                case 3: { // victory
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->victory = static_cast<bool>(value);
                    break;
                }
                case 4: { // exp_gained
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->exp_gained = static_cast<int32_t>(value);
                    break;
                }
                case 5: { // gold_gained
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->gold_gained = static_cast<int32_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillTokenBalanceResponse message
struct BillTokenBalanceResponse {
    std::string user_id;
    std::string token_type;
    int64_t balance;
    int64_t locked;

    // Default constructor
    BillTokenBalanceResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize user_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize token_type
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        // Serialize balance
        buffer.push_back(24); // Field key: 24 (field number 3, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(balance);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize locked
        buffer.push_back(32); // Field key: 32 (field number 4, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(locked);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillTokenBalanceResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillTokenBalanceResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // balance
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->balance = static_cast<int64_t>(value);
                    break;
                }
                case 4: { // locked
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->locked = static_cast<int64_t>(value);
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

// BillPaymentResponse message
struct BillPaymentResponse {
    std::string order_id;
    std::string user_id;
    int64_t amount;
    std::string currency;
    std::string token_type;
    int64_t token_amount;
    std::string payment_method;
    std::string status;
    std::string transaction_id;
    std::string payment_url;

    // Default constructor
    BillPaymentResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize order_id
        buffer.push_back(10); // Field key: 10 (field number 1, wire type 2)
            // String encoding
            uint64_t length = order_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), order_id.begin(), order_id.end());

        // Serialize user_id
        buffer.push_back(18); // Field key: 18 (field number 2, wire type 2)
            // String encoding
            uint64_t length = user_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), user_id.begin(), user_id.end());

        // Serialize amount
        buffer.push_back(24); // Field key: 24 (field number 3, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize currency
        buffer.push_back(34); // Field key: 34 (field number 4, wire type 2)
            // String encoding
            uint64_t length = currency.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), currency.begin(), currency.end());

        // Serialize token_type
        buffer.push_back(42); // Field key: 42 (field number 5, wire type 2)
            // String encoding
            uint64_t length = token_type.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), token_type.begin(), token_type.end());

        // Serialize token_amount
        buffer.push_back(48); // Field key: 48 (field number 6, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(token_amount);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize payment_method
        buffer.push_back(58); // Field key: 58 (field number 7, wire type 2)
            // String encoding
            uint64_t length = payment_method.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), payment_method.begin(), payment_method.end());

        // Serialize status
        buffer.push_back(66); // Field key: 66 (field number 8, wire type 2)
            // String encoding
            uint64_t length = status.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), status.begin(), status.end());

        // Serialize transaction_id
        buffer.push_back(74); // Field key: 74 (field number 9, wire type 2)
            // String encoding
            uint64_t length = transaction_id.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), transaction_id.begin(), transaction_id.end());

        // Serialize payment_url
        buffer.push_back(82); // Field key: 82 (field number 10, wire type 2)
            // String encoding
            uint64_t length = payment_url.size();
            while (length > 127) {
                buffer.push_back(static_cast<uint8_t>((length & 0x7F) | 0x80));
                length >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(length));
            buffer.insert(buffer.end(), payment_url.begin(), payment_url.end());

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BillPaymentResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BillPaymentResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 1: { // order_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->order_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 2: { // user_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->user_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 3: { // amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->amount = static_cast<int64_t>(value);
                    break;
                }
                case 4: { // currency
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->currency = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 5: { // token_type
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->token_type = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 6: { // token_amount
                    // Varint decoding
                    uint64_t value = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        value |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        value |= buffer[offset++] << shift;
                    }
                    msg->token_amount = static_cast<int64_t>(value);
                    break;
                }
                case 7: { // payment_method
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->payment_method = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 8: { // status
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->status = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 9: { // transaction_id
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->transaction_id = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                case 10: { // payment_url
                    // String decoding
                    uint64_t length = 0;
                    int shift = 0;
                    while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                        length |= (buffer[offset++] & 0x7F) << shift;
                        shift += 7;
                    }
                    if (offset < buffer.size()) {
                        length |= buffer[offset++] << shift;
                    }
                    if (offset + length <= buffer.size()) {
                        msg->payment_url = std::string(buffer.begin() + offset, buffer.begin() + offset + length);
                        offset += length;
                    }
                    break;
                }
                default:
                    // Skip unknown fields
                    if (wireType == 0) { // Varint
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) ++offset;
                        if (offset < buffer.size()) ++offset;
                    } else if (wireType == 2) { // Length-delimited
                        uint64_t length = 0;
                        int shift = 0;
                        while (offset < buffer.size() && (buffer[offset] & 0x80)) {
                            length |= (buffer[offset++] & 0x7F) << shift;
                            shift += 7;
                        }
                        if (offset < buffer.size()) {
                            length |= buffer[offset++] << shift;
                        }
                        offset += length;
                    } else if (wireType == 5) { // 32-bit
                        offset += 4;
                    } else if (wireType == 1) { // 64-bit
                        offset += 8;
                    }
                    break;
            }
        }

        return msg;
    }
};

} // namespace game

#endif // GAME_H
