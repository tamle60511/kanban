package translate

func TranslateReport(report map[string]interface{}) map[string]interface{} {
	translatedReport := make(map[string]interface{})
	for key, value := range report {
		if translatedValue, ok := enToVnTranslate[key]; ok {
			translatedReport[translatedValue] = value
		}
	}
	return translatedReport
}

func TranslateKey(key string) string {
	if translatedValue, ok := enToVnTranslate[key]; ok {
		return translatedValue
	}
	return key
}
