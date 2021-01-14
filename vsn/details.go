package vsn

import (
	"encoding/hex"

	"github.com/superp00t/gophercraft/bnet/realmlist"
)

type BuildInfo struct {
	MajorVersion    uint32
	MinorVersion    uint32
	BugfixVersion   uint32
	HotfixVersion   string
	WinAuthSeed     []byte
	Win64AuthSeed   []byte
	Mac64AuthSeed   []byte
	WinChecksumSeed []byte
	MacChecksumSeed []byte
}

// This data will be displayed in the JSON/protobuf realm listing.
var details = map[Build]*BuildInfo{}

// https://github.com/TrinityCore/TrinityCore/blob/master/sql/base/auth_database.sql#L470
func init() {
	addBuild(3368, 0, 5, 3, "", "", "", "", "", "")
	addBuild(5875, 1, 12, 1, "", "", "", "", "95EDB27C7823B363CBDDAB56A392E7CB73FCCA20", "8D173CC381961EEBABF336F5E6675B101BB513E5")
	addBuild(6005, 1, 12, 2, "", "", "", "", "", "")
	addBuild(6141, 1, 12, 3, "", "", "", "", "", "")
	addBuild(8606, 2, 4, 3, "", "", "", "", "319AFAA3F2559682F9FF658BE01456255F456FB1", "D8B0ECFE534BC1131E19BAD1D4C0E813EEE4994F")
	addBuild(9947, 3, 1, 3, "", "", "", "", "", "")
	addBuild(10505, 3, 2, 2, "a", "", "", "", "", "")
	addBuild(11159, 3, 3, 0, "a", "", "", "", "", "")
	addBuild(11403, 3, 3, 2, "", "", "", "", "", "")
	addBuild(11723, 3, 3, 3, "a", "", "", "", "", "")
	addBuild(12340, 3, 3, 5, "a", "", "", "", "CDCBBD5188315E6B4D19449D492DBCFAF156A347", "B706D13FF2F4018839729461E3F8A0E2B5FDC034")
	addBuild(13623, 4, 0, 6, "a", "", "", "", "", "")
	addBuild(13930, 3, 3, 5, "a", "", "", "", "", "")
	addBuild(14545, 4, 2, 2, "", "", "", "", "", "")
	addBuild(15595, 4, 3, 4, "", "", "", "", "", "")
	addBuild(19116, 6, 0, 3, "", "", "", "", "", "")
	addBuild(19243, 6, 0, 3, "", "", "", "", "", "")
	addBuild(19342, 6, 0, 3, "", "", "", "", "", "")
	addBuild(19702, 6, 1, 0, "", "", "", "", "", "")
	addBuild(19802, 6, 1, 2, "", "", "", "", "", "")
	addBuild(19831, 6, 1, 2, "", "", "", "", "", "")
	addBuild(19865, 6, 1, 2, "", "", "", "", "", "")
	addBuild(20182, 6, 2, 0, "a", "", "", "", "", "")
	addBuild(20201, 6, 2, 0, "", "", "", "", "", "")
	addBuild(20216, 6, 2, 0, "", "", "", "", "", "")
	addBuild(20253, 6, 2, 0, "", "", "", "", "", "")
	addBuild(20338, 6, 2, 0, "", "", "", "", "", "")
	addBuild(20444, 6, 2, 2, "", "", "", "", "", "")
	addBuild(20490, 6, 2, 2, "a", "", "", "", "", "")
	addBuild(20574, 6, 2, 2, "a", "", "", "", "", "")
	addBuild(20726, 6, 2, 3, "", "", "", "", "", "")
	addBuild(20779, 6, 2, 3, "", "", "", "", "", "")
	addBuild(20886, 6, 2, 3, "", "", "", "", "", "")
	addBuild(21355, 6, 2, 4, "", "", "", "", "", "")
	addBuild(21463, 6, 2, 4, "", "", "", "", "", "")
	addBuild(21742, 6, 2, 4, "", "", "", "", "", "")
	addBuild(22248, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22293, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22345, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22410, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22423, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22498, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22522, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22566, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22594, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22624, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22747, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22810, 7, 0, 3, "", "", "", "", "", "")
	addBuild(22900, 7, 1, 0, "", "", "", "", "", "")
	addBuild(22908, 7, 1, 0, "", "", "", "", "", "")
	addBuild(22950, 7, 1, 0, "", "", "", "", "", "")
	addBuild(22995, 7, 1, 0, "", "", "", "", "", "")
	addBuild(22996, 7, 1, 0, "", "", "", "", "", "")
	addBuild(23171, 7, 1, 0, "", "", "", "", "", "")
	addBuild(23222, 7, 1, 0, "", "", "", "", "", "")
	addBuild(23360, 7, 1, 5, "", "", "", "", "", "")
	addBuild(23420, 7, 1, 5, "", "", "", "", "", "")
	addBuild(23911, 7, 2, 0, "", "", "", "", "", "")
	addBuild(23937, 7, 2, 0, "", "", "", "", "", "")
	addBuild(24015, 7, 2, 0, "", "", "", "", "", "")
	addBuild(24330, 7, 2, 5, "", "", "", "", "", "")
	addBuild(24367, 7, 2, 5, "", "", "", "", "", "")
	addBuild(24415, 7, 2, 5, "", "", "", "", "", "")
	addBuild(24430, 7, 2, 5, "", "", "", "", "", "")
	addBuild(24461, 7, 2, 5, "", "", "", "", "", "")
	addBuild(24742, 7, 2, 5, "", "", "", "", "", "")
	addBuild(25549, 7, 3, 2, "", "FE594FC35E7F9AFF86D99D8A364AB297", "1252624ED8CBD6FAC7D33F5D67A535F3", "66FC5E09B8706126795F140308C8C1D8", "", "")
	addBuild(25996, 7, 3, 5, "", "23C59C5963CBEF5B728D13A50878DFCB", "C7FF932D6A2174A3D538CA7212136D2B", "210B970149D6F56CAC9BADF2AAC91E8E", "", "")
	addBuild(26124, 7, 3, 5, "", "F8C05AE372DECA1D6C81DA7A8D1C5C39", "46DF06D0147BA67BA49AF553435E093F", "C9CA997AB8EDE1C65465CB2920869C4E", "", "")
	addBuild(26365, 7, 3, 5, "", "2AAC82C80E829E2CA902D70CFA1A833A", "59A53F307288454B419B13E694DF503C", "DBE7F860276D6B400AAA86B35D51A417", "", "")
	addBuild(26654, 7, 3, 5, "", "FAC2D693E702B9EC9F750F17245696D8", "A752640E8B99FE5B57C1320BC492895A", "9234C1BD5E9687ADBD19F764F2E0E811", "", "")
	addBuild(26822, 7, 3, 5, "", "283E8D77ECF7060BE6347BE4EB99C7C7", "2B05F6D746C0C6CC7EF79450B309E595", "91003668C245D14ECD8DF094E065E06B", "", "")
	addBuild(26899, 7, 3, 5, "", "F462CD2FE4EA3EADF875308FDBB18C99", "3551EF0028B51E92170559BD25644B03", "8368EFC2021329110A16339D298200D4", "", "")
	addBuild(26972, 7, 3, 5, "", "797ECC19662DCBD5090A4481173F1D26", "6E212DEF6A0124A3D9AD07F5E322F7AE", "341CFEFE3D72ACA9A4407DC535DED66A", "", "")
	addBuild(28153, 8, 0, 1, "", "", "DD626517CC6D31932B479934CCDC0ABF", "", "", "")
	addBuild(30706, 8, 1, 5, "", "", "BB6D9866FE4A19A568015198783003FC", "", "", "")
	addBuild(30993, 8, 2, 0, "", "", "2BAD61655ABC2FC3D04893B536403A91", "", "", "")
	addBuild(31229, 8, 2, 0, "", "", "8A46F23670309F2AAE85C9A47276382B", "", "", "")
	addBuild(31429, 8, 2, 0, "", "", "7795A507AF9DC3525EFF724FEE17E70C", "", "", "")
	addBuild(31478, 8, 2, 0, "", "", "7973A8D54BDB8B798D9297B096E771EF", "", "", "")
	addBuild(32305, 8, 2, 5, "", "", "21F5A6FC7AD89FBF411FDA8B8738186A", "", "", "")
	addBuild(32494, 8, 2, 5, "", "", "58984ACE04919401835C61309A848F8A", "", "", "")
	addBuild(32580, 8, 2, 5, "", "", "87C2FAA0D7931BF016299025C0DDCA14", "", "", "")
	addBuild(32638, 8, 2, 5, "", "", "5D07ECE7D4A867DDDE615DAD22B76D4E", "", "", "")
	addBuild(32722, 8, 2, 5, "", "", "1A09BE1D38A122586B4931BECCEAD4AA", "", "", "")
	addBuild(32750, 8, 2, 5, "", "", "C5CB669F5A5B237D1355430877173207", "EF1F4E4D099EA2A81FD4C0DEBC1E7086", "", "")
	addBuild(32978, 8, 2, 5, "", "", "76AE2EA03E525D97F5688843F5489000", "1852C1F847E795D6EB45278CD433F339", "", "")
	addBuild(33369, 8, 3, 0, "", "", "5986AC18B04D3C403F56A0CF8C4F0A14", "F5A849C70A1054F07EA3AB833EBF6671", "", "")
	addBuild(33528, 8, 3, 0, "", "", "0ECE033CA9B11D92F7D2792C785B47DF", "", "", "")
	addBuild(33724, 8, 3, 0, "", "", "38F7BBCF284939DD20E8C64CDBF9FE77", "", "", "")
	addBuild(33775, 8, 3, 0, "", "", "B826300A8449ED0F6EF16EA747FA2D2E", "354D2DE619D124EE1398F76B0436FCFC", "", "")
	addBuild(33941, 8, 3, 0, "", "", "88AF1A36D2770D0A6CA086497096A889", "", "", "")
	addBuild(34220, 8, 3, 0, "", "", "B5E35B976C6BAF82505700E7D9666A2C", "", "", "")
	addBuild(34601, 8, 3, 0, "", "", "0D7DF38F725FABA4F009257799A10563", "", "", "")
	addBuild(34769, 8, 3, 0, "", "", "93F9B9AF6397E3E4EED94D36D16907D2", "", "", "")
	addBuild(34963, 8, 3, 0, "", "", "7BA50C879C5D04221423B02AC3603A11", "C5658A17E702163447BAAAE46D130A1B", "", "")
	addBuild(35249, 8, 3, 7, "", "", "C7B11F9AE9FF1409F5582902B3D10D1C", "", "", "")
	addBuild(35284, 8, 3, 7, "", "", "EA3818E7DCFD2009DBFC83EE3C1E4F1B", "A6201B0AC5A73D13AB2FDCC79BB252AF", "", "")
	addBuild(35435, 8, 3, 7, "", "", "BB397A92FE23740EA52FC2B5BA2EC8E0", "8FE657C14A46BCDB2CE6DA37E430450E", "", "")
	addBuild(35662, 8, 3, 7, "", "", "578BC94870C278CB6962F30E6DC203BB", "5966016C368ED9F7AAB603EE6703081C", "", "")
}

func addBuild(build Build, major, minor, bugfix uint32, hotfix string, winAuthSeed, win64AuthSeed, mac64AuthSeed, winChecksumSeed, macChecksumSeed string) {
	bi := &BuildInfo{}
	bi.MajorVersion = major
	bi.MinorVersion = minor
	bi.BugfixVersion = bugfix
	bi.HotfixVersion = hotfix

	var err error
	if winAuthSeed != "" {
		bi.WinAuthSeed, err = hex.DecodeString(winAuthSeed)
		if err != nil {
			panic(err)
		}
	}
	if win64AuthSeed != "" {
		bi.Win64AuthSeed, err = hex.DecodeString(win64AuthSeed)
		if err != nil {
			panic(err)
		}
	}

	if mac64AuthSeed != "" {
		bi.Mac64AuthSeed, err = hex.DecodeString(mac64AuthSeed)
		if err != nil {
			panic(err)
		}
	}

	if winChecksumSeed != "" {
		bi.WinChecksumSeed, err = hex.DecodeString(winChecksumSeed)
		if err != nil {
			panic(err)
		}
	}

	if macChecksumSeed != "" {
		bi.MacChecksumSeed, err = hex.DecodeString(macChecksumSeed)
		if err != nil {
			panic(err)
		}
	}

	details[build] = bi
}

// (legacy auth protocol) Returns a 3-byte field describing the version data. For instance, version 3.3.5 would return []byte{ 3, 3, 5 }
func (b Build) VersionData() []byte {
	d := b.ClientVersion()
	return []byte{
		uint8(*d.VersionMajor),
		uint8(*d.VersionMinor),
		uint8(*d.VersionRevision),
	}
}

func (b Build) ClientVersion() *realmlist.ClientVersion {
	info := b.BuildInfo()
	if info == nil {
		return nil
	}

	vsn := new(realmlist.ClientVersion)

	__build := uint32(b)

	vsn.VersionBuild = &__build
	vsn.VersionMajor = &info.MajorVersion
	vsn.VersionMinor = &info.MinorVersion
	vsn.VersionRevision = &info.BugfixVersion

	return vsn
}

func (b Build) BuildInfo() *BuildInfo {
	return details[b]
}
