package update

func init32305() {
	dc := NewDescriptorCompiler(32305)

	obj := dc.ObjectBase()
	obj.GUID(ObjectGUID, Public)
}
