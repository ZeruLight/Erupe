from hexdump import hexdump
import io
import sys

from construct import *

Binary8Header = Struct(
    "server_type" / Bytes(3),
    "entry_count" / Int16ub,
    "body_size" / Int16ub,
    "checksum" / Int32ub,
)

EntranceListComplete = Struct(
    Embedded(Binary8Header),
    "servers" / Array(this.entry_count,
        Struct(
            "host_ip_4byte" / Int32ub,
            "unk_1" / Int16ub, # Server ID maybe?
            "unk_2" / Int16ub,
            "channel_count" / Int16ub,
            "server_type" / Byte, # Server type. 0=?, 1=open, 2=cities, 3=newbie, 4=bar
            "color" / Byte, # Server activity. 0 = green, 1 = orange, 2 = blue
            "unk_6" / Byte, # Something to do with server recommendation on 0, 3, and 5.
            "name" / Bytes(66), # Shift-JIS.

            # 4096(PC, PS3/PS4)?, 8258(PC, PS3/PS4)?, 8192 == nothing?
            # THIS ONLY EXISTS IF Binary8Header.type == "SV2", NOT "SVR"!
            "allowed_client_type_flags" / Int32ub,

            "channels" / Array(this.channel_count,
                Struct(
                    "port" / Int16ub,
                    "unk_1" / Int16ub, # Channel ID maybe?
                    "max_players" / Int16ub,
                    "current_players" / Int16ub,
                    "unk_4" / Int16ub,
                    "unk_5" / Int16ub,
                    "unk_6" / Int16ub,
                    "unk_7" / Int16ub,
                    "unk_8" / Int16ub,
                    "unk_9" / Int16ub,
                    "unk_10" / Int16ub,
                    "unk_11" / Int16ub,
                    "unk_12" / Int16ub,
                    "unk_13" / Int16ub,
                )
            ),
        )
    ),
)


BINARY8_KEY = bytes([0x01, 0x23, 0x34, 0x45, 0x56, 0xAB, 0xCD, 0xEF])
def decode_binary8(data, unk_key_byte):
    cur_key = ((54323 * unk_key_byte) + 1) & 0xFFFFFFFF

    output_data = bytearray()
    for i in range(len(data)):
        tmp = (data[i] ^ (cur_key >> 13)) & 0xFF
        output_data.append(tmp ^ BINARY8_KEY[i&7])
        cur_key = ((54323 * cur_key) + 1) & 0xFFFFFFFF

    return output_data

def encode_binary8(data, unk_key_byte):
    cur_key = ((54323 * unk_key_byte) + 1) & 0xFFFFFFFF

    output_data = bytearray()
    for i in range(len(data)):
        output_data.append(data[i] ^ (BINARY8_KEY[i&7] ^ ((cur_key >> 13) & 0xFF)))
        cur_key = ((54323 * cur_key) + 1) & 0xFFFFFFFF

    return output_data


SUM32_TABLE_0 = bytes([0x35, 0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12])
SUM32_TABLE_1 = bytes([0x7A, 0xAA, 0x97, 0x53, 0x66, 0x12, 0xDE, 0xDE, 0x35])
def calc_sum32(data):
    length = len(data)

    t0_i = length & 0xFF
    t1_i = data[length >> 1]

    out = bytearray(4)
    for i in range(len(data)):
        t0_i += 1
        t1_i += 1

        tmp = (SUM32_TABLE_1[t1_i % 9] ^ SUM32_TABLE_0[t0_i % 7]) ^ data[i]
        out[i & 3] = (out[i & 3] + tmp) & 0xFF

    return Int32ub.parse(out)

def read_binary8_part(stream):
    # Read the header and decrypt the header first to get the size.
    enc_bytes = bytearray(stream.read(12))
    dec_header_bytes = decode_binary8(enc_bytes[1:], enc_bytes[0])
    header = Binary8Header.parse(dec_header_bytes)

    # Then read the body, append to the header, and decrypt the full thing.
    body_bytes = stream.read(header.body_size)
    enc_bytes.extend(body_bytes)
    dec_bytes = decode_binary8(enc_bytes[1:], enc_bytes[0])

    # Then return the parsed header and just the raw body data.
    return (enc_bytes[0], header, dec_bytes[11:], dec_bytes)

def write_binary8_part(key, server_type, entry_count, payload):
    body = Binary8Header.build(dict(
        server_type=server_type,
        entry_count=entry_count,
        body_size=len(payload),
        checksum=calc_sum32(payload),
    ))

    temp = bytearray()
    temp.extend(body)
    temp.extend(payload)

    out = bytearray()
    out.append(key)
    out.extend(encode_binary8(temp, key))

    return out


def pad_bytes_to_len(b, length):
    out = bytearray(b)
    diff = length-len(out)
    out.extend(bytearray(diff))
    return bytes(out)



def make_custom_entrance_server_resp():
    # Get the userinfo_data
    with open('tw_server_list_resp.bin', 'rb') as f:
        (key, header, data, raw_dec_bytes) = read_binary8_part(f)
        userinfo_data = f.read()
        hexdump(userinfo_data)

    server_count = 1

    data = EntranceListComplete.build(dict(
        server_type = b'SV2',
        entry_count = server_count,
        body_size = 0xFFFF,
        checksum = 0xFFFFFFFF,
        servers = [dict(
            host_ip_4byte = 0x0100007F, #0x7F000001,#3377555739,
            unk_1 = 16,
            unk_2 = 0,
            channel_count = 1,
            server_type = 1,
            color = 0, # Server activity. 0 = green, 1 = orange, 2 = blue
            unk_6 = 3,
            name = pad_bytes_to_len("AErupe Server @localhost".encode('shift-jis'), 66),
            allowed_client_type_flags = 4096, # 4096(PC, PS3/PS4)?, 8192 == nothing?
            channels = [dict(
                port = 54001,
                unk_1 = 16,
                max_players = 100,
                current_players = 0,
                unk_4 = 0,
                unk_5 = 0,
                unk_6 = 0,
                unk_7 = 0,
                unk_8 = 0,
                unk_9 = 0,
                unk_10 = 319,
                unk_11 = 248,#254,
                unk_12 = 159,#255,
                unk_13 = 12345
            )],
        )]
    ))

    print(data)

    reencoded = write_binary8_part(0, b'SV2', server_count, data[11:])
    with open('custom_entrance_server_resp.bin', 'wb') as f:
        f.write(reencoded)
        f.write(userinfo_data)


    with open('custom_entrance_server_resp.bin', 'rb') as f:
        (key, header, data, raw_dec_bytes) = read_binary8_part(f)
        print(EntranceListComplete.parse(raw_dec_bytes[0:]))




make_custom_entrance_server_resp()


"""
with open('tw_server_list_resp.bin', 'rb') as f:
    (key, header, data, raw_dec_bytes) = read_binary8_part(f)
    print(EntranceListComplete.parse(raw_dec_bytes[0:]))
"""

"""
with open('tw_server_list_resp.bin', 'rb') as f:
    filedata = f.read()

    rdr = io.BytesIO(filedata)

    (key, header, data, raw_dec_bytes) = read_binary8_part(rdr)
    userinfo_data = rdr.read()

    reencoded = write_binary8_part(key, header.server_type, header.entry_count, data)


    hexdump(reencoded[:16])
    hexdump(filedata[:16])

"""

"""

with open('dec_bin8_data_dump.bin', 'rb') as f:
    print("calc_sum32: {:X}".format(calc_sum32(f.read())))
    print("want: 74EF4928")
"""