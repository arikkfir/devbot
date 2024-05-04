package configuration

// import (
// 	"fmt"
// 	"github.com/jessevdk/go-flags"
// 	"github.com/rs/zerolog/log"
// 	"github.com/spf13/pflag"
// 	"os"
// 	"reflect"
// 	"unicode"
// )
//
// func Parse(cfg interface{}) {
// 	configType := reflect.TypeOf(cfg)
// 	configValue := reflect.ValueOf(cfg)
// 	for i := 0; i < configType.NumField(); i++ {
// 		field := configType.Field(i)
// 		fieldType := field.Type
// 		fieldValue := configValue.Field(i)
// 		key := camelCaseToDash(field.Name)
// 		description := field.Tag.Get("desc")
//
// 		switch fieldType.Kind() {
// 		case reflect.Bool:
// 			pflag.BoolVar(fieldValue.Addr().Interface().(*bool), key, fieldValue.Bool(), description)
// 			ApplyBoolEnvironmentVariableTo(fieldValue.Addr().Interface().(*bool), FlagNameToEnvironmentVariable(key))
// 		case reflect.Int:
// 			pflag.IntVar(fieldValue.Addr().Interface().(*int), key, int(fieldValue.Int()), description)
// 		case reflect.Int8:
// 			pflag.Int8Var(fieldValue.Addr().Interface().(*int8), key, int8(fieldValue.Int()), description)
// 		case reflect.Int16:
// 			pflag.Int16Var(fieldValue.Addr().Interface().(*int16), key, int16(fieldValue.Int()), description)
// 		case reflect.Int32:
// 			pflag.Int32Var(fieldValue.Addr().Interface().(*int32), key, int32(fieldValue.Int()), description)
// 		case reflect.Int64:
// 			pflag.Int64Var(fieldValue.Addr().Interface().(*int64), key, fieldValue.Int(), description)
// 		case reflect.Uint:
// 			pflag.UintVar(fieldValue.Addr().Interface().(*uint), key, uint(fieldValue.Uint()), description)
// 		case reflect.Uint8:
// 			pflag.Uint8Var(fieldValue.Addr().Interface().(*uint8), key, uint8(fieldValue.Uint()), description)
// 		case reflect.Uint16:
// 			pflag.Uint16Var(fieldValue.Addr().Interface().(*uint16), key, uint16(fieldValue.Uint()), description)
// 		case reflect.Uint32:
// 			pflag.Uint32Var(fieldValue.Addr().Interface().(*uint32), key, uint32(fieldValue.Uint()), description)
// 		case reflect.Uint64:
// 			pflag.Uint64Var(fieldValue.Addr().Interface().(*uint64), key, fieldValue.Uint(), description)
// 		case reflect.Float32:
// 			pflag.Float32Var(fieldValue.Addr().Interface().(*float32), key, float32(fieldValue.Float()), description)
// 		case reflect.Float64:
// 			pflag.Float64Var(fieldValue.Addr().Interface().(*float64), key, fieldValue.Float(), description)
// 		case reflect.Slice:
// 			switch fieldType.Elem().Kind() {
// 			case reflect.Bool:
// 				pflag.BoolSliceVar(fieldValue.Addr().Interface().(*[]bool), key, fieldValue.Interface().([]bool), description)
// 			case reflect.Int:
// 				pflag.IntSliceVar(fieldValue.Addr().Interface().(*[]int), key, fieldValue.Interface().([]int), description)
// 			case reflect.Int32:
// 				pflag.Int32SliceVar(fieldValue.Addr().Interface().(*[]int32), key, fieldValue.Interface().([]int32), description)
// 			case reflect.Int64:
// 				pflag.Int64SliceVar(fieldValue.Addr().Interface().(*[]int64), key, fieldValue.Interface().([]int64), description)
// 			case reflect.Uint:
// 				pflag.UintSliceVar(fieldValue.Addr().Interface().(*[]uint), key, fieldValue.Interface().([]uint), description)
// 			case reflect.Float32:
// 				pflag.Float32SliceVar(fieldValue.Addr().Interface().(*[]float32), key, fieldValue.Interface().([]float32), description)
// 			case reflect.Float64:
// 				pflag.Float64SliceVar(fieldValue.Addr().Interface().(*[]float64), key, fieldValue.Interface().([]float64), description)
// 			case reflect.String:
// 				pflag.StringSliceVar(fieldValue.Addr().Interface().(*[]string), key, fieldValue.Interface().([]string), description)
// 			default:
// 				log.Fatal().Msgf("Unsupported configuration type: %s", fieldType)
// 			}
// 		case reflect.String:
// 			pflag.StringVar(fieldValue.Addr().Interface().(*string), key, fieldValue.String(), description)
// 		default:
// 			log.Fatal().Msgf("Unsupported configuration type: %s", fieldType)
// 		}
// 	}
//
// 	parser := flags.NewParser(cfg, flags.HelpFlag|flags.PassDoubleDash)
// 	if _, err := parser.Parse(); err != nil {
// 		fmt.Printf("ERROR: %s\n\n", err)
// 		parser.WriteHelp(os.Stderr)
// 		os.Exit(1)
// 	}
// }
//
// func camelCaseToDash(camelCaseStr string) string {
// 	var result []rune
// 	for i, r := range camelCaseStr {
// 		if unicode.IsUpper(r) {
// 			if i != 0 {
// 				result = append(result, '-')
// 			}
// 			result = append(result, unicode.ToLower(r))
// 		} else {
// 			result = append(result, r)
// 		}
// 	}
// 	return string(result)
// }
