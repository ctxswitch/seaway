package client

func ObjectKeyFromObject(obj Object) ObjectKey {
	return ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}
