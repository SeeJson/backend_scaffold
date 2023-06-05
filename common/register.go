package common

import "github.com/SeeJson/backend_scaffold/apispec"

func RegisterEnumSpec(ab *apispec.ApiSpecBuilder) {
	for k, ens := range enumStrings {
		for _, en := range ens {
			ab.AddEnum(k, apispec.NewEnum(en.k, 0, en.v))
		}
	}
	for k, ens := range allEnums {
		for _, en := range ens {
			ab.AddEnum(k, apispec.NewEnum(en.e.Name, en.e.Int, en.k))
		}
	}

}
