package address

/*
func (this *Address) UpdateCache() error {
	err := cache_service.Set(fmt.Sprintf("%s_%d", TABLE, this.ID()), this.GetData())
	if err != nil {
		return err
	}
	return nil
}


func GetWithCache(id int64) (*Address, error) {
	obj := New()
	err := cache_service.Load(fmt.Sprintf("%s_%d", TABLE, id), obj)
	if err != nil {
		return nil, err
	}
	if !tools.Empty(obj) {
		return obj, nil
	}

	obj, err = Get(id)
	if err != nil {
		return nil, err
	}
	err = obj.UpdateCache()
	if err != nil {
		return nil, err
	}

	return obj, nil
}
*/
