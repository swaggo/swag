package api

import "github.com/swaggo/swag/testdata/enums/types"

//	 post students
//
//		@Summary      test enums in response models
//		@Description  test enums in response models
//		@Failure      400   {object}  types.Person  "ok"
//		@Router       /students [post]
func API() {
	_ = types.Person{}
}

//	 get students
//
//		@Summary      test enums in response request
//		@Description  test enums in response request
//		@Param 		  typeinquery query []types.Type true "type"
//		@Param 		  typeinheader header types.Type true "type"
//		@Param 		  typeinpath path types.Type true "type"
//		@Success      200   "ok"
//		@Router       /students/{typeinpath}/ [get]
func API2() {
	_ = types.Person{}
}

//	 post students
//
//		@Summary      test enums fields in formdata request
//		@Description  test enums fields in formdata request
//		@Param 		  student formData types.Person true "type"
//		@Success      200   "ok"
//		@Router       /students2 [get]
func API3() {
	_ = types.Person{}
}

//	 post students
//
//		@Summary      test array enums fields in formdata request
//		@Description  test array enums fields in formdata request
//		@Param 		  student formData types.PersonWithArrayEnum true "type"
//		@Success      200   "ok"
//		@Router       /students4 [get]
func API4() {
	_ = types.Person{}
}
