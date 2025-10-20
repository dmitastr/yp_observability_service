package models

// DisplayMetric used to represent metrics on web page and store value as string
type DisplayMetric struct {
	Name        string
	Type        string
	StringValue string
}

// ModelToDisplay converts [models.Metrics] to [models.DisplayMetric] and converts metric value to string
func ModelToDisplay(m Metrics) DisplayMetric {
	val, err := m.GetValueString()
	if err != nil {
		val = ""
	}
	return DisplayMetric{Name: m.ID, Type: m.MType, StringValue: val}
}
