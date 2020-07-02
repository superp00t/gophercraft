package packet

// type Parser struct {
// 	Version uint32
// }

// func NewParser(version uint32) *Parser {
// 	return &Parser{version}
// }

// type MultiPacket struct {
// 	Data []struct {
// 		Opcode WorldType
// 		Data   []byte
// 	}
// }

// const cmpConstant = 0x9827D8F1

// func UnmarshalCompressedPacket(b []byte) (WorldType, []byte, error) {
// 	expected := len(b) - 12

// 	e := etc.FromBytes(b)
// 	uncompressedSize := e.ReadUint32()
// 	uncompressedAdler := e.ReadUint32()
// 	// compressedAdler := e.ReadUint32()
// 	e.ReadUint32()
// 	fmt.Println("expecting", expected)
// 	compressedData := etc.FromBytes(e.ReadBytes(e.Available()))
// 	if compressedData.Len() != expected {
// 		panic(compressedData.Len())
// 	}

// 	if uncompressedSize > 120000 {
// 		return 0, nil, fmt.Errorf("packet: attempted to decompress extremely large packet.")
// 	}

// 	// pCompressedAdler := Adler32(cmpConstant, compressedData.Bytes())

// 	// if pCompressedAdler != compressedAdler {
// 	// 	return 0, nil, fmt.Errorf("packet: UnmarshalCompressedPacket: compressed packet checksum mismatch (packet's 0x%08X !== calculated 0x%08X)", compressedAdler, pCompressedAdler)
// 	// }

// 	z, err := zlib.NewReader(compressedData)
// 	if err != nil {
// 		return 0, nil, err
// 	}

// 	// unc := make([]byte, uncompressedSize)
// 	// _, err = z.Read(unc)
// 	// if err != nil {
// 	// 	return 0, nil, err
// 	// }

// 	unc, err := ioutil.ReadAll(z)
// 	if err != nil {

// 	}

// 	z.Close()

// 	pUncompressedAdler := Adler32(
// 		Adler32(cmpConstant, unc[:2]),
// 		unc[2:],
// 	)

// 	if pUncompressedAdler != uncompressedAdler {
// 		return 0, nil, fmt.Errorf("packet: UnmarshalCompressedPacket: uncompressed packet checksum mismatch")
// 	}

// 	u16 := binary.LittleEndian.Uint16(unc[:2])

// 	return WorldType(u16), unc[2:], nil
// }

// func UnmarshalMultiPacket(b []byte) (*MultiPacket, error) {
// 	mp := new(MultiPacket)
// 	e := etc.FromBytes(b)
// 	for e.Available() > 0 {
// 		ln := e.ReadUint16()
// 		opcode := WorldType(e.ReadUint16())
// 		if e.Available() < int(ln) {
// 			return nil, fmt.Errorf("packet: UnmarshalMultiPacket, unexpected EOF")
// 		}
// 		data := e.ReadBytes(int(ln))
// 		mp.Data = append(mp.Data, struct {
// 			Opcode WorldType
// 			Data   []byte
// 		}{opcode, data})
// 	}
// 	return mp, nil
// }

// type Content struct {
// 	Type        WorldType
// 	Description string
// 	Bytes       []byte
// 	Data        interface{}
// }

// func (p *Parser) Parse(smsg bool, opcode WorldType, data []byte) ([]Content, error) {
// 	switch opcode {
// 	case M_SMSG_MULTIPLE_PACKETS:
// 		dat, err := UnmarshalMultiPacket(data)
// 		if err != nil {
// 			return nil, err
// 		}

// 		var pc []Content

// 		for _, v := range dat.Data {
// 			ct, err := p.Parse(smsg, v.Opcode, v.Data)
// 			if err != nil {
// 				return nil, err
// 			}

// 			pc = append(pc, ct...)
// 		}

// 		return pc, nil
// 	case M_SMSG_COMPRESSED_PACKET:
// 		return []Content{{
// 			Type:        M_SMSG_COMPRESSED_PACKET,
// 			Description: "server sent a compressed packet",
// 			Bytes:       nil,
// 			Data:        nil,
// 		}}, nil
// 	case SMSG_COMPRESSED_UPDATE_OBJECT:
// 		e := etc.FromBytes(data)
// 		e.ReadUint32()
// 		z, err := zlib.NewReader(e)
// 		if err != nil {
// 			return nil, err
// 		}

// 		o := etc.NewBuffer()
// 		io.Copy(o, e)
// 		z.Close()

// 		return p.parseUpdateObject(opcode, o.Bytes())
// 	case SMSG_UPDATE_OBJECT:
// 		return p.parseUpdateObject(opcode, data)
// 	case CMSG_CHAR_ENUM:
// 		return emptyDescription(opcode, data, "client requested character list")
// 	case SMSG_CHAR_ENUM:
// 		i, err := UnmarshalCharacterList(p.Version, data)
// 		return []Content{{
// 			opcode,
// 			"server sent char list",
// 			data,
// 			i,
// 		}}, err
// 	case SMSG_WARDEN_DATA:
// 		return emptyDescription(opcode, data, "server requested Warden anticheat data")
// 	case CMSG_WARDEN_DATA:
// 		return emptyDescription(opcode, data, "client uploaded Warden data")
// 	default:
// 		return []Content{{
// 			opcode,
// 			fmt.Sprintf("unknown purpose (%d bytes)", len(data)),
// 			data,
// 			nil,
// 		}}, nil
// 	}
// }

// func emptyDescription(wt WorldType, data []byte, desc string) ([]Content, error) {
// 	return []Content{{
// 		wt,
// 		desc,
// 		data,
// 		nil,
// 	}}, nil
// }

// func (p *Parser) GUIDDisplayString(g guid.GUID) string {
// 	return g.String()
// }

// func (p *Parser) parseUpdateObject(wt WorldType, data []byte) ([]Content, error) {
// 	uo, err := update.Unmarshal(p.Version, data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	description := ""

// 	for _, v := range uo.Blocks {
// 		switch v.Data.Type() {
// 		case update.CreateObject, update.SpawnObject:
// 			description += fmt.Sprintf("created gameobject %s.\n", p.GUIDDisplayString(v.GUID))
// 		case update.Values:
// 			vb := v.Data.(*update.ValuesBlock)
// 			if len(vb.Values) == 1 {
// 				for k := range vb.Values {
// 					description += fmt.Sprintf("updated %s value: %s\n", p.GUIDDisplayString(v.GUID), k.String())
// 				}
// 			} else {
// 				description += fmt.Sprintf("updated %s %d values\n", p.GUIDDisplayString(v.GUID), len(vb.Values))
// 			}
// 		}
// 	}

// 	return []Content{{
// 		wt,
// 		description,
// 		data,
// 		uo,
// 	}}, nil
// }
