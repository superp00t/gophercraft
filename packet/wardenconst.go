package packet

/*
 * Copyright (C) 2008-2017 TrinityCore <http://www.trinitycore.org/>
 * Copyright (C) 2005-2011 MaNGOS <http://getmangos.com/>
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the
 * Free Software Foundation; either version 2 of the License or (at your
 * option) any later version.
 *
 * This program is distributed in the hope that it will be useful but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program. If not see <http://www.gnu.org/licenses/>.
 */

const (

	// Client->Server
	WARDEN_CMSG_MODULE_MISSING      uint8 = 0
	WARDEN_CMSG_MODULE_OK           uint8 = 1
	WARDEN_CMSG_CHEAT_CHECKS_RESULT uint8 = 2
	WARDEN_CMSG_MEM_CHECKS_RESULT   uint8 = 3 // only sent if MEM_CHECK bytes doesn't match
	WARDEN_CMSG_HASH_RESULT         uint8 = 4
	WARDEN_CMSG_MODULE_FAILED       uint8 = 5 // this is sent when client failed to load uploaded module due to cache fail

	// Server->Client
	WARDEN_SMSG_MODULE_USE           uint8 = 0
	WARDEN_SMSG_MODULE_CACHE         uint8 = 1
	WARDEN_SMSG_CHEAT_CHECKS_REQUEST uint8 = 2
	WARDEN_SMSG_MODULE_INITIALIZE    uint8 = 3
	WARDEN_SMSG_MEM_CHECKS_REQUEST   uint8 = 4 // byte len; while (!EOF) { byte unk(1); byte index(++); string module(can be 0); int offset; byte len; byte[] bytes_to_compare[len]; }
	WARDEN_SMSG_HASH_REQUEST         uint8 = 5

	MEM_CHECK     uint8 = 0xF3 // 243: byte moduleNameIndex + uint Offset + byte Len (check to ensure memory isn't modified)
	PAGE_CHECK_A  uint8 = 0xB2 // 178: uint Seed + byte[20] SHA1 + uint Addr + byte Len (scans all pages for specified hash)
	PAGE_CHECK_B  uint8 = 0xBF // 191: uint Seed + byte[20] SHA1 + uint Addr + byte Len (scans only pages starts with MZ+PE headers for specified hash)
	MPQ_CHECK     uint8 = 0x98 // 152: byte fileNameIndex (check to ensure MPQ file isn't modified)
	LUA_STR_CHECK uint8 = 0x8B // 139: byte luaNameIndex (check to ensure LUA string isn't used)
	DRIVER_CHECK  uint8 = 0x71 // 113: uint Seed + byte[20] SHA1 + byte driverNameIndex (check to ensure driver isn't loaded)
	TIMING_CHECK  uint8 = 0x57 //  87: empty (check to ensure GetTickCount() isn't detoured)
	PROC_CHECK    uint8 = 0x7E // 126: uint Seed + byte[20] SHA1 + byte moluleNameIndex + byte procNameIndex + uint Offset + byte Len (check to ensure proc isn't detoured)
	MODULE_CHECK  uint8 = 0xD9 // 217: uint Seed + byte[20] SHA1 (check to ensure module isn't injected)
)
