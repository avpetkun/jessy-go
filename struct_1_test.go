package jessy

import (
	"encoding/json"
	"reflect"
	"testing"
	"unsafe"

	"github.com/avpetkun/jessy-go/zgo"
)

func TestStruct1(t *testing.T) {
	//nilText := json.RawMessage(nil)
	rawText := json.RawMessage([]byte(`"123"`))

	println("rawText", rawText, "rawText addr", &rawText)

	println("------------")
	str1 := struct{ M *json.RawMessage }{&rawText}
	println("str1.M", str1.M, "str1.M addr", &str1.M)

	eface := zgo.UnpackEface(str1)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str2 := struct {
		M *json.RawMessage
		X int
	}{&rawText, 123}
	println("str2.M", str2.M, "str2.M addr", &str2.M, "str2.X addr", &str2.X)

	eface = zgo.UnpackEface(str2)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str3 := struct{ M json.RawMessage }{rawText}
	println("str3.M", str3.M, "str3.M addr", &str3.M)

	eface = zgo.UnpackEface(str3)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str4 := struct {
		M json.RawMessage
		X int
	}{rawText, 123}
	println("str4.M", str4.M, "str4.M addr", &str4.M, "str4.X addr", &str4.X)

	eface = zgo.UnpackEface(str4)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	//
	//

	println("------------")
	str11 := &struct{ M *json.RawMessage }{&rawText}
	println("str11", str11, "str11.M", str11.M, "str11.M addr", &str11.M)

	eface = zgo.UnpackEface(str11)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str13 := &struct{ M json.RawMessage }{rawText}
	println("str13", str13, "str13.M", str13.M, "str13.M addr", &str13.M)

	eface = zgo.UnpackEface(str13)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str22 := &struct {
		M *json.RawMessage
		X int
	}{&rawText, 123}
	println("str22", str22, "str22.M", str22.M, "str22.M addr", &str22.M, "str22.X addr", &str22.X)

	eface = zgo.UnpackEface(str22)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))

	println("------------")
	str23 := &struct {
		X int
		M *json.RawMessage
	}{123, &rawText}
	println("str23", str23, "str23.M", str23.M, "str23.M addr", &str23.M, "str23.X addr", &str23.X)

	eface = zgo.UnpackEface(str23)
	eface.Value = unsafe.Add(eface.Value, 8)
	println("eface value", eface.Value, "unpacked", *(*unsafe.Pointer)(eface.Value))
}

func TestStruct2(t *testing.T) {
	nilText := json.RawMessage(nil)

	println("------------")
	str0 := struct{ M *json.RawMessage }{&nilText}
	println(`struct{ M *json.RawMessage }{&nilText}`, "str0.M", str0.M, "str0.M addr", &str0.M)
	findM("str0", str0)

	println("------------")
	str00 := &struct{ M *json.RawMessage }{&nilText}
	println(`&struct{ M *json.RawMessage }{&nilText}`, "str00.M", str00.M, "str00.M addr", &str00.M)
	findM("str00", str00)

	println("------------")
	str000 := &struct{ M json.RawMessage }{nilText}
	println(`&struct{ M json.RawMessage }{nilText}`, "str000.M", str000.M, "str000.M addr", &str000.M)
	findM("str000", str000)

	println("------------")
	str0000 := struct{ M json.RawMessage }{nilText}
	println(`struct{ M json.RawMessage }{nilText}`, "str0000.M", str0000.M, "str0000.M addr", &str0000.M)
	findM("str0000", str0000)

	println("------------")
	rawText := json.RawMessage([]byte(`"123"`))

	println("rawText", rawText, "rawText addr", &rawText)

	println("------------")
	str1 := struct{ M *json.RawMessage }{&rawText}
	println(`struct{ M *json.RawMessage }{&rawText}`, "str1.M", str1.M, "str1.M addr", &str1.M)
	findM("str1", str1)

	println("------------")
	str3 := struct{ M json.RawMessage }{rawText}
	println("str3.M", str3.M, "str3.M addr", &str3.M)
	findM("str3", str3)

	println("------------")
	str11 := &struct{ M *json.RawMessage }{&rawText}
	println("str11", str11, "str11.M", str11.M, "str11.M addr", &str11.M)
	findM("str11", str11)

	println("------------")
	str13 := &struct{ M json.RawMessage }{rawText}
	println("str13", str13, "str13.M", str13.M, "str13.M addr", &str13.M)
	findM("str13", str13)

	println("------------")
	str133 := &struct{ M json.RawMessage }{}
	println("str133", str133, "str133.M", str133.M, "str133.M addr", &str133.M)
	findM("str133", str133)

	println("------------")
	str134 := &struct{ M *json.RawMessage }{}
	println("str134", str134, "str134.M", str134.M, "str134.M addr", &str134.M)
	findM("str134", str134)

	println("------------")
	str135 := struct{ M *json.RawMessage }{}
	println("str135", &str135, "str135.M", str135.M, "str135.M addr", &str135.M)
	findM("str135", str135)

	println("------------")
	str136 := struct{ M json.RawMessage }{}
	println("str136", &str136, "str136.M", str136.M, "str136.M addr", &str136.M)
	findM("str136", str136)

	println("------------")
	str2 := struct {
		M *json.RawMessage
		X int
	}{&rawText, 123}
	println(`struct { M *json.RawMessage, X int }{&rawText, 123}`, "str2.M", str2.M, "str2.M addr", &str2.M, "str2.X addr", &str2.X)
	findM("str2", str2)

	println("------------")
	str4 := struct {
		M json.RawMessage
		X int
	}{rawText, 123}
	println("str4.M", str4.M, "str4.M addr", &str4.M, "str4.X addr", &str4.X)
	findM("str4", str4)

	println("------------")
	str5 := struct {
		X int
		M json.RawMessage
	}{123, rawText}
	println("str5.M", str5.M, "str5.M addr", &str5.M, "str5.X addr", &str5.X)
	findM("str5", str5)

	println("------------")
	str22 := &struct {
		M *json.RawMessage
		X int
	}{&rawText, 123}
	println("str22", str22, "str22.M", str22.M, "str22.M addr", &str22.M, "str22.X addr", &str22.X)
	findM("str22", str22)

	println("------------")
	str23 := &struct {
		X int
		M *json.RawMessage
	}{123, &rawText}
	println("str23", str23, "str23.M", str23.M, "str23.M addr", &str23.M, "str23.X addr", &str23.X)
	findM("str23", str23)

	println("------------")
	str24 := &struct {
		X int
		M *json.RawMessage
	}{123, nil}
	println("str24", str24, "str24.M", str24.M, "str24.M addr", &str24.M, "str24.X addr", &str24.X)
	findM("str24", str24)

	println("------------")
	com1 := &struct {
		X int
		Z struct{ json.RawMessage }
	}{
		X: 123,
		Z: struct{ json.RawMessage }{nilText},
	}
	findM("com1", com1)

	println("------------")
	com11 := &struct {
		X int
		Z struct{ json.RawMessage }
	}{
		X: 123,
		Z: struct{ json.RawMessage }{rawText},
	}
	findM("com11", com11)

	println("------------")
	com2 := &struct {
		X int
		Z struct{ *json.RawMessage }
	}{
		X: 123,
		Z: struct{ *json.RawMessage }{&rawText},
	}
	findM("com2", com2)

	println("------------")
	com3 := &struct {
		X int
		Z *struct{ *json.RawMessage }
	}{
		X: 123,
		Z: &struct{ *json.RawMessage }{&nilText},
	}
	findM("com3", com3)

	println("------------")
	com4 := &struct {
		X int
		Z *struct{ json.RawMessage }
	}{
		X: 123,
		Z: &struct{ json.RawMessage }{rawText},
	}
	findM("com4", com4)

	println("------------")
	com5 := &struct {
		X int
		Z *struct {
			Y int
			J json.RawMessage
		}
	}{
		X: 123,
		Z: &struct {
			Y int
			J json.RawMessage
		}{123, rawText},
	}
	findM("com5", com5)

	println("------------")
	com6 := &struct {
		X int
		Z struct {
			Y int
			J json.RawMessage
		}
	}{
		X: 123,
		Z: struct {
			Y int
			J json.RawMessage
		}{123, rawText},
	}
	findM("com6", com6)

	println("------------")
	num := 12345
	com7 := &struct {
		X int
		D *int
		Z struct {
			Y int
			W *int
			J json.RawMessage
		}
	}{
		X: 123,
		D: &num,
		Z: struct {
			Y int
			W *int
			J json.RawMessage
		}{123, &num, rawText},
	}
	findM("com7", com7)

	println("------------")
	com8 := struct {
		X int
		D *int
		Z *struct {
			Y int
			W *int
			J json.RawMessage
		}
	}{
		X: 123,
		D: &num,
		Z: &struct {
			Y int
			W *int
			J json.RawMessage
		}{
			Y: 123,
			W: &num,
			J: json.RawMessage{},
		},
	}
	findM("com8", com8)

	println("------------")
	com9 := struct {
		X  int
		D  *int
		DD **int
	}{
		X: 123,
		D: &num,
	}
	com9.DD = &com9.D
	findM("com9", com9)
}

func findM(name string, value any) {
	eface := zgo.UnpackEface(value)
	findDeep(name, eface.Type.Native(), eface.Value, false, false)
}

func findDeep(name string, t reflect.Type, value unsafe.Pointer, byPointer, doUnpack bool) {
	if doUnpack {
		value = *(*unsafe.Pointer)(value)
	}
	if value == nil {
		return
	}
	switch t.Kind() {
	case reflect.Pointer:
		doUnpack = t.Elem().Kind() != reflect.Struct
		findDeep(name, t.Elem(), value, true, doUnpack)
	case reflect.Struct:
		for i := range t.NumField() {
			f := t.Field(i)
			ft := f.Type
			doUnpack = false
			if t.NumField() == 1 {
				if ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
					doUnpack = byPointer
				}
				byPointer = false
			} else {
				byPointer = ft.Kind() == reflect.Struct
				if !byPointer && ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
					doUnpack = true
				}
			}
			fValue := unsafe.Add(value, f.Offset)
			findDeep(name, ft, fValue, byPointer, doUnpack)
		}
	case reflect.Int:
		println(name, "> X", value, "->", *(*int)(value))
	case reflect.Slice:
		b := *(*[]byte)(value)
		if len(b) > 100 {
			panic("bad slice")
		}
		println(name, "> M", value, "->", string(b))
	}
}

func findMOld(name string, value any) {
	eface := zgo.UnpackEface(value)

	t := eface.Type.Native()
	if t.Kind() == reflect.Struct && t.NumField() == 1 {
		findMValue(name, t.Field(0).Type, eface.Value)
		return
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	for i := range t.NumField() {
		f := t.Field(i)
		if f.Name == "M" || f.Anonymous {
			if f.Type.Kind() == reflect.Pointer {
				findMValue(name, f.Type, *(*unsafe.Pointer)(unsafe.Add(eface.Value, f.Offset)))
			} else {
				findMValue(name, f.Type, unsafe.Add(eface.Value, f.Offset))
			}
		}
	}
}

func findMValue(name string, typ reflect.Type, value unsafe.Pointer) {
	println(name, ">", typ.String(), value, "->", *(*[]byte)(value))
}
