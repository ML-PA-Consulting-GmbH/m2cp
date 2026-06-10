package format

//func TestPrettyMarshallString(t *testing.T) {
//	var err error
//	data := "test"
//	bytes, err := Marshal(data)
//	assert.NoError(t, err)
//	assert.Equal(t, []byte{}, bytes)
//}
//
//func TestPrettyMarshallSimpleStruct(t *testing.T) {
//	var err error
//	type Pair struct {
//		Key   string `cli:"key"`
//		Value string `cli:"value"`
//	}
//	var data struct {
//		DefinitionList []Pair `cli:"section,list"`
//	}
//	data.DefinitionList = []Pair{
//		{
//			"k1",
//			"v1",
//		},
//		{
//			"k2",
//			"v2",
//		},
//	}
//
//	bytes, err := Marshal(data)
//	assert.NoError(t, err)
//	assert.Equal(t, []byte{}, bytes)
//
//	// TODO: shall be:
//	// section:
//	// k1: v1
//	// k2: v2
//}
//
//func TestPrettyMarshallSimpleTable(t *testing.T) {
//	var err error
//	type Pair struct {
//		Key   string `cli:"key"`
//		Value string `cli:"value"`
//	}
//	var data struct {
//		DefinitionList []Pair `cli:"table,table"`
//	}
//	data.DefinitionList = []Pair{
//		{
//			"k1",
//			"v1",
//		},
//		{
//			"k2",
//			"v2",
//		},
//		{
//			"k3",
//			"v3",
//		},
//	}
//
//	bytes, err := Marshal(data)
//	assert.NoError(t, err)
//	assert.Equal(t, []byte{}, bytes)
//
//	// TODO: shall be:
//	// table:
//	// k1 k2 k3
//	// v1 v2 v3
//
//	//bytes, err := Marshal(data)
//	//assert.NoError(t, err)
//	//assert.Equal(t, []byte{}, bytes)
//}
//
//func TestPrettyMarshallCombination(t *testing.T) {
//	var err error
//	type Pair struct {
//		Key   string `cli:"key"`
//		Value string `cli:"value"`
//	}
//	var data struct {
//		DefinitionList []Pair `cli:"definitionList,listing"`
//		Table          []Pair `cli:"table,table"`
//	}
//	data.DefinitionList = []Pair{
//		{
//			"k1",
//			"v1",
//		},
//		{
//			"k2",
//			"v2",
//		},
//	}
//	data.Table = []Pair{
//		{
//			"k3",
//			"v3",
//		},
//		{
//			"k4",
//			"v4",
//		},
//		{
//			"k5",
//			"v5",
//		},
//	}
//
//	bytes, err := Marshal(data)
//	assert.NoError(t, err)
//	assert.Equal(t, []byte{}, bytes)
//
//	// TODO: shall be:
//	// definitionList:
//	// k1: v1
//	// k2: v2
//	// table:
//	// k3 k4 k5
//	// v3 v4 v5
//
//	//bytes, err := Marshal(data)
//	//assert.NoError(t, err)
//	//assert.Equal(t, []byte{}, bytes)
//}
