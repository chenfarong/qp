#ifndef BAG_H
#define BAG_H

#include <cstdint>
#include <string>
#include <vector>
#include <memory>
#include <stdexcept>

namespace game {

// BagItemData message
struct BagItemData {
    int64_t item_id;
    int32_t item_cfg_id;
    int64_t num;

    // Default constructor
    BagItemData() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize item_id
        buffer.push_back(0); // Field key: 0 (field number 0, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(item_id);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize item_cfg_id
        buffer.push_back(0); // Field key: 0 (field number 0, wire type 0)
            // Varint encoding for int32
            uint64_t value = static_cast<uint64_t>(item_cfg_id);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        // Serialize num
        buffer.push_back(0); // Field key: 0 (field number 0, wire type 0)
            // Varint encoding for int64
            uint64_t value = static_cast<uint64_t>(num);
            while (value > 127) {
                buffer.push_back(static_cast<uint8_t>((value & 0x7F) | 0x80));
                value >>= 7;
            }
            buffer.push_back(static_cast<uint8_t>(value));

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BagItemData> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BagItemData>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 0: { // item_id
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
                    msg->item_id = static_cast<int64_t>(value);
                    break;
                }
                case 0: { // item_cfg_id
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
                    msg->item_cfg_id = static_cast<int32_t>(value);
                    break;
                }
                case 0: { // num
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
                    msg->num = static_cast<int64_t>(value);
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

// BagRequest message
struct BagRequest {

    // Default constructor
    BagRequest() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        return buffer;
    }

    // Deserialize from byte array
    static std::unique_ptr<BagRequest> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BagRequest>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
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

// BagResponse message
struct BagResponse {
    std::vector<BagItemData> items;

    // Default constructor
    BagResponse() = default;

    // Serialize to byte array
    std::vector<uint8_t> Serialize() const {
        std::vector<uint8_t> buffer;
        size_t offset = 0;

        // Serialize items (repeated)
        for (const auto& item : items) {
            // Field key: 2 (field number 0, wire type 2)
            buffer.push_back(2);
            // Nested message encoding for BagItemData
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
    static std::unique_ptr<BagResponse> Deserialize(const std::vector<uint8_t>& buffer) {
        auto msg = std::make_unique<BagResponse>();
        size_t offset = 0;

        while (offset < buffer.size()) {
            uint8_t key = buffer[offset++];
            int fieldNumber = key >> 3;
            int wireType = key & 0x07;

            switch (fieldNumber) {
                case 0: { // items
                    // Nested message decoding for BagItemData
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
                        msg->items.push_back(*BagItemData::Deserialize(nestedData));
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

#endif // BAG_H
