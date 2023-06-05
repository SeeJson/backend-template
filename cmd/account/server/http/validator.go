package httpserver

import (
	"reflect"
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
)

const (
	patternPhone = `^(0|\+?86)?` + // 匹配 0,86,+86
		`(13[0-9]|` + // 130-139
		`14[57]|` + // 145,147
		`15[0-35-9]|` + // 150-153,155-159
		`17[0678]|` + // 170,176,177,17u
		`18[0-9])` + // 180-189
		`[0-9]{8}$`
)

var (
	regPhone = regexp.MustCompile("^" + patternPhone + "$")
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}

	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// add any custom validations etc. here
		v.validate.RegisterValidation("phone", phoneValidator)
	})
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func phoneValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return regPhone.MatchString(phone)
}
