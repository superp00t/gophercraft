package pkzip

func decompress_pkzip(in_buf *[]byte, in_size uint32, out_buf *[]byte, out_size uint32) (int, error) {
	inf := new(info)
	inf.in_buf = in_buf
	inf.in_bytes = int32(in_size)
	inf.out_buf = out_buf
	inf.max_out = int32(out_size)

	var err error
	err = do_decompress_pkzip(inf)
	if err != nil {
		return 0, err
	}

	return int(inf.out_pos), nil
}

func Decompress(input []byte) ([]byte, error) {
	inp := input
	out := make([]byte, len(inp)*10)
	sz, err := decompress_pkzip(&inp, uint32(len(inp)), &out, uint32(len(inp)))
	if err != nil {
		return nil, err
	}

	return out[:sz], nil
}
