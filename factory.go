package fixtory

import (
	"reflect"
	"testing"
)

type BluePrintFunc func(i int, last interface{}) interface{}

type Factory struct {
	t           *testing.T
	productType reflect.Type
	// struct
	last interface{}
	// index is the next struct index (which is equal to already built struct count in this factory)
	index int
	// v is a pointer to struct
	OnBuild func(t *testing.T, v interface{})
	// v is a pointer to struct
	OnInsert func(t *testing.T, v interface{})
}

type Builder struct {
	*Factory
	// index is the next struct index in this builder
	index           int
	bluePrint       func(i int, last interface{}) interface{}
	traitValues     []interface{}
	traitZeroes     [][]string
	eachParamValues []interface{}
	eachParamZeroes [][]string
	setValues       interface{}
	zeroFields      []string
	resetAfterBuild bool
}

func NewFactory(t *testing.T, v interface{}) *Factory {
	return &Factory{t: t, productType: reflect.PtrTo(reflect.TypeOf(v)), index: 0, last: v}
}

func (f *Factory) NewBuilder(bluePrint BluePrintFunc, traitValues []interface{}, traitZeroes [][]string) *Builder {
	return &Builder{Factory: f, bluePrint: bluePrint, traitValues: traitValues, traitZeroes: traitZeroes}
}

func (f *Factory) Reset() {
	f.last = reflect.New(f.productType.Elem()).Elem().Interface()
	f.index = 0
}

func (b *Builder) EachParam(values []interface{}, zeroes [][]string) *Builder {
	b.eachParamValues = values
	b.eachParamZeroes = zeroes
	return b
}

func (b *Builder) Set(v interface{}) *Builder {
	b.setValues = v
	return b
}

func (b *Builder) Zero(fields ...string) *Builder {
	b.zeroFields = fields
	return b
}

func (b *Builder) ResetAfter() *Builder {
	b.resetAfterBuild = true
	return b
}

func (b *Builder) Build() interface{} {
	b.index = 0
	product := b.build(false)
	if b.resetAfterBuild {
		b.Factory.Reset()
	}
	return product
}

func (b *Builder) BuildList(n int) []interface{} {
	b.index = 0
	products := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		products = append(products, b.build(false))
	}
	if b.resetAfterBuild {
		b.Factory.Reset()
	}
	return products
}

func (b *Builder) build(insert bool) interface{} {
	product := reflect.New(b.productType.Elem()).Interface()

	if b.bluePrint != nil {
		MapNotZeroFields(b.bluePrint(b.Factory.index, b.last), product)
	}
	for i, trait := range b.traitValues {
		// Map the non-zero fields in each trait value struct, and also set the zero
		// fields for each trait sequentially.
		MapNotZeroFields(trait, product)
		for _, f := range b.traitZeroes[i] {
			uf := reflect.ValueOf(product).Elem().FieldByName(f)
			uf.Set(reflect.Zero(uf.Type()))
		}
	}
	if len(b.eachParamValues) > b.index {
		// Map the non-zero fields in each trait value struct, and also set the zero
		// fields for each trait sequentially.
		MapNotZeroFields(b.eachParamValues[b.index], product)
		for _, f := range b.eachParamZeroes[b.index] {
			uf := reflect.ValueOf(product).Elem().FieldByName(f)
			uf.Set(reflect.Zero(uf.Type()))
		}
	}
	if b.setValues != nil {
		MapNotZeroFields(b.setValues, product) // map the non-zero overridden values
	}
	for _, f := range b.zeroFields {
		uf := reflect.ValueOf(product).Elem().FieldByName(f)
		uf.Set(reflect.Zero(uf.Type()))
	}

	b.last = reflect.ValueOf(product).Elem().Interface()
	b.index++
	b.Factory.index++

	if b.OnBuild != nil {
		b.OnBuild(b.t, product)
	}
	if insert && b.OnInsert != nil {
		b.OnInsert(b.t, product)
	}
	return product
}

func (b *Builder) Insert() interface{} {
	b.index = 0
	product := b.build(true)
	if b.resetAfterBuild {
		b.Factory.Reset()
	}
	return product
}

func (b *Builder) InsertList(n int) []interface{} {
	b.index = 0
	products := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		products = append(products, b.build(true))
	}
	if b.resetAfterBuild {
		b.Factory.Reset()
	}
	return products
}
