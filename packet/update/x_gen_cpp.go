package update

// func (dc *DescriptorCompiler) GenerateCPP() (string, error) {
// 	str := etc.NewBuffer()

// 	for _, v := range dc.Classes {
// 		fmt.Fprintf(str, "Class* %s = ", v.Name)
// 		if v.Extends != nil {
// 			fmt.Fprintf(str, "%s->extend()", v.Extends.Name)
// 		} else {
// 			fmt.Fprintf(str, "ObjectBase();")
// 		}
// 		fmt.Fprintf(str, "\n")
// 		for _, field := range v.Fields {
// 			fmt.Fprintf(str, "%s->%s(%s) // abs offset = 0x%04X  relative = 0x%04X\n", v.Name, field.FieldType.String(), field.Global, field.AbsBlockOffset(), field.BlockOffset)
// 		}
// 		fmt.Fprintf(str, "\n\n")
// 	}

// 	return str.ToString(), nil
// }
