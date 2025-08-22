package parser

func TurnLinesIntoObjectsV2(profilePath string) ([]*LineObj, error) {
	profileData, err := extractProfileData(profilePath)
	if err != nil {
		return nil, err
	}

	var lineObjs []*LineObj
	for _, entry := range profileData.SortedEntries {
		fn := entry.Name
		lineObj := &LineObj{
			FnName:         fn,
			Flat:           float64(entry.Flat),
			FlatPercentage: profileData.FlatPercentages[fn],
			SumPercentage:  profileData.SumPercentages[fn],
			Cum:            float64(profileData.Cum[fn]),
			CumPercentage:  profileData.CumPercentages[fn],
		}
		lineObjs = append(lineObjs, lineObj)
	}

	return lineObjs, nil
}
