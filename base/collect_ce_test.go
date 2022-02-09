package base

import (
	"testing"
)

func TestRunCollectCE(t *testing.T) {
	ct := &collectCETask{
		TaskMeta: TaskMeta{
			Task: StringPointer(CollectCETaskName),
		},
		With: collectCEInputs{
			CESysdigToken: StringPointer("eyJraWQiOiIyMDIyMDExNjA4MjIiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC01NTAwMDBLTUc4IiwiaWQiOiJJQk1pZC01NTAwMDBLTUc4IiwicmVhbG1pZCI6IklCTWlkIiwic2Vzc2lvbl9pZCI6IkMtMzQ2YWUyNWUtNjYwZS00ZGRhLWFlYjgtN2NjYzU0NTdmODRkIiwianRpIjoiOGUxZjYxNmMtY2FhNS00ZDk2LTgzZmUtZjgwNmVjMjIwZTdhIiwiaWRlbnRpZmllciI6IjU1MDAwMEtNRzgiLCJnaXZlbl9uYW1lIjoiQUxBTiIsImZhbWlseV9uYW1lIjoiQ0hBIiwibmFtZSI6IkFMQU4gQ0hBIiwiZW1haWwiOiJBbGFuLkNoYTFAaWJtLmNvbSIsInN1YiI6IkFsYW4uQ2hhMUBpYm0uY29tIiwiYXV0aG4iOnsic3ViIjoiQWxhbi5DaGExQGlibS5jb20iLCJpYW1faWQiOiJJQk1pZC01NTAwMDBLTUc4IiwibmFtZSI6IkFMQU4gQ0hBIiwiZ2l2ZW5fbmFtZSI6IkFMQU4iLCJmYW1pbHlfbmFtZSI6IkNIQSIsImVtYWlsIjoiQWxhbi5DaGExQGlibS5jb20ifSwiYWNjb3VudCI6eyJib3VuZGFyeSI6Imdsb2JhbCIsInZhbGlkIjp0cnVlLCJic3MiOiJjODc1ZGM2NTI1NDBkMThmZWRhZDY3ZjBmYzU3MTVmOCIsImltc191c2VyX2lkIjoiOTA4NTk4OCIsImltcyI6IjE1MjkwMzkifSwiaWF0IjoxNjQ0NDI4NTM3LCJleHAiOjE2NDQ0Mjk3MzcsImlzcyI6Imh0dHBzOi8vaWFtLmNsb3VkLmlibS5jb20vaWRlbnRpdHkiLCJncmFudF90eXBlIjoidXJuOmlibTpwYXJhbXM6b2F1dGg6Z3JhbnQtdHlwZTpwYXNzY29kZSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImJ4IiwiYWNyIjoxLCJhbXIiOlsicHdkIl19.kh91dl19CJpZeQ79CuI2m0yrZMdq4OAPq5wHuc7i2JDvxiIFflUI91xteAGU-ISxvzkIaB7EbMueHKeKurzJk336LLl6mWftf_K3sxL1BVmy_JqhgRckeopNgtUuhN-Zpmb5UUqUNguQuh4EyJ78XkcceTQpzHl49c2MHVVlTEWoAGLd8t6wuKrgJMugWYf1Lyd-GxISal1-Cvt1GwozLIuhzQfl_7VmypZlBUa4PVjlVmcvRm0-WcaVbSHC4uLqe_fgZiz3w1YDnHabB_68F45-TluCckCOucTSjzVcnnK7vtLAI2oPACbFUoG64vXFFDarJQuRdaxGI7ZxeQX8dw"),
			IBMInstanceId: StringPointer("cee3ac9a-15cf-45a5-9321-a52d4c5cc97e"),
			URL:           StringPointer("https://ca-tor.monitoring.cloud.ibm.com/api/data"),
		},
	}

	exp := &Experiment{
		Tasks:  []Task{ct},
		Result: &ExperimentResult{},
	}
	exp.InitResults()
	ct.Run(exp)

	// err := ct.Run(exp)
	// assert.NoError(t, err)
	// assert.Equal(t, exp.Result.Insights.NumVersions, 1)
}
