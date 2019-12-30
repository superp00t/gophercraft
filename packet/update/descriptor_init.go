package update

var (
	Descriptors = map[uint32]*DescriptorCompiler{}
)

func init() {
	// vanilla (1.12.1)
	init5875()
}
