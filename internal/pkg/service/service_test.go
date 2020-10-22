package service_test

// var data *service.Data

// func initTest(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 1}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"xxxx", "mama"}, {"xxxx", "."}}, {{"xxx", "."}}}}
// 	data = newTestData(&testTagger{res: tr}, &testLex{res: sr})
// }

// func TestNotFound(t *testing.T) {
// 	initTest(t)
// 	req, err := http.NewRequest("GET", "/any", nil)
// 	assert.Nil(t, err)
// 	resp := httptest.NewRecorder()

// 	service.NewRouter(data).ServeHTTP(resp, req)
// 	assert.Equal(t, resp.Code, http.StatusNotFound)
// }

// func TestProvides(t *testing.T) {
// 	initTest(t)
// 	req, err := http.NewRequest("POST", "/tag", newTestInput("mama o"))
// 	assert.Nil(t, err)
// 	resp := httptest.NewRecorder()

// 	service.NewRouter(data).ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusOK, resp.Code)
// 	assert.Equal(t, `[{"type":"WORD","string":"mama","mi":"mama","lemma":"xxxx"},{"type":"SPACE","string":" "},{"type":"WORD","string":"o","mi":".","lemma":"xxx"}]`,
// 		strings.TrimSpace(resp.Body.String()))
// }

// func TestFailsPreprocess(t *testing.T) {
// 	initTest(t)
// 	req, err := http.NewRequest("POST", "/tag", newTestInput("mama o"))
// 	assert.Nil(t, err)
// 	resp := httptest.NewRecorder()

// 	data.Preprocesor = &testPreprocessor{err: errors.New("err")}
// 	service.NewRouter(data).ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusInternalServerError, resp.Code)
// }

// func TestFailsMorph(t *testing.T) {
// 	initTest(t)
// 	req, err := http.NewRequest("POST", "/tag", newTestInput("mama o"))
// 	assert.Nil(t, err)
// 	resp := httptest.NewRecorder()

// 	data.Tagger = &testTagger{err: errors.New("err")}
// 	service.NewRouter(data).ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusInternalServerError, resp.Code)
// }

// func TestFailsLex(t *testing.T) {
// 	initTest(t)
// 	req, err := http.NewRequest("POST", "/tag", newTestInput("mama o"))
// 	assert.Nil(t, err)
// 	resp := httptest.NewRecorder()

// 	data.Segmenter = &testLex{err: errors.New("err")}
// 	service.NewRouter(data).ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusInternalServerError, resp.Code)
// }

// func TestMapOK(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}}}
// 	r, err := service.MapRes("mami", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, len(r))
// 	assert.Equal(t, "WORD", r[0].Type)
// 	assert.Equal(t, "mami", r[0].String)
// 	assert.Equal(t, "mama", r[0].Lemma)
// 	assert.Equal(t, "xxxx", r[0].Mi)
// }

// func TestMapSeveral(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 2}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}, {{"oo", "xoo"}}}}
// 	r, err := service.MapRes("mami oi", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(r))
// 	assert.Equal(t, "WORD", r[0].Type)
// 	assert.Equal(t, "mami", r[0].String)
// 	assert.Equal(t, "mama", r[0].Lemma)
// 	assert.Equal(t, "xxxx", r[0].Mi)

// 	assert.Equal(t, "SPACE", r[1].Type)
// 	assert.Equal(t, " ", r[1].String)

// 	assert.Equal(t, "WORD", r[2].Type)
// 	assert.Equal(t, "oi", r[2].String)
// 	assert.Equal(t, "oo", r[2].Lemma)
// 	assert.Equal(t, "xoo", r[2].Mi)
// }

// func TestMapUTF(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 2}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"mama", "xxxx"}}, {{"oo", "xoo"}}}}
// 	r, err := service.MapRes("mamą oš", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(r))
// 	assert.Equal(t, "WORD", r[0].Type)
// 	assert.Equal(t, "mamą", r[0].String)
// 	assert.Equal(t, "mama", r[0].Lemma)
// 	assert.Equal(t, "xxxx", r[0].Mi)

// 	assert.Equal(t, "SPACE", r[1].Type)

// 	assert.Equal(t, "WORD", r[2].Type)
// 	assert.Equal(t, "oš", r[2].String)
// 	assert.Equal(t, "oo", r[2].Lemma)
// 	assert.Equal(t, "xoo", r[2].Mi)
// }

// func TestMapSep(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{".", "T."}}}}
// 	r, err := service.MapRes(".", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, len(r))
// 	assert.Equal(t, "SEPARATOR", r[0].Type)
// 	assert.Equal(t, ".", r[0].String)
// 	assert.Equal(t, "", r[0].Lemma)
// 	assert.Equal(t, "T.", r[0].Mi)
// }

// func TestMapSpace(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 1}, {6, 1}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{".", "T."}},{{".", "T."}}}}
// 	r, err := service.MapRes(".  \n \n.", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(r))
// 	assert.Equal(t, "SPACE", r[1].Type)
// 	assert.Equal(t, "  \n \n", r[1].String)
// }

// func TestMapNumber(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
// 	r, err := service.MapRes("1234", tr, sr)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, len(r))
// 	assert.Equal(t, "NUMBER", r[0].Type)
// 	assert.Equal(t, "1234", r[0].String)
// 	assert.Equal(t, "M----d-", r[0].Mi)
// }

// func TestMapErrTooLongSeg(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
// 	_, err := service.MapRes("123", tr, sr)
// 	assert.NotNil(t, err)
// }

// func TestMapErrWrongSeg(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, -1}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"1234", "M----d-"}}}}
// 	_, err := service.MapRes("123", tr, sr)
// 	assert.NotNil(t, err)
// 	sr = &api.SegmenterResult{Seg: [][]int{{0, 0}}}
// 	_, err = service.MapRes("123", tr, sr)
// 	assert.NotNil(t, err)
// 	sr = &api.SegmenterResult{Seg: [][]int{{0}}}
// 	_, err = service.MapRes("123", tr, sr)
// 	assert.NotNil(t, err)
// 	sr = &api.SegmenterResult{Seg: [][]int{{0, 1}, {0, 2}}}
// 	_, err = service.MapRes("1234", tr, sr)
// 	assert.NotNil(t, err)
// }

// func TestMapErrWrongMorph(t *testing.T) {
// 	sr := &api.SegmenterResult{Seg: [][]int{{0, 4}}}
// 	tr := &api.TaggerResult{Msd: [][][]string{{{"1234"}}}}
// 	_, err := service.MapRes("1234", tr, sr)
// 	assert.NotNil(t, err)
// 	tr = &api.TaggerResult{Msd: [][][]string{{{}}}}
// 	_, err = service.MapRes("1234", tr, sr)
// 	assert.NotNil(t, err)
// 	sr = &api.SegmenterResult{Seg: [][]int{{0, 4}, {5, 1}}}
// 	tr = &api.TaggerResult{Msd: [][][]string{{{"1234", "xx"}}}}
// 	_, err = service.MapRes("1234 .", tr, sr)
// 	assert.NotNil(t, err)
// }

// func newTestInput(text string) *bytes.Buffer {
// 	result := new(bytes.Buffer)
// 	result.WriteString(text)
// 	return result
// }

// type testTagger struct {
// 	res *api.TaggerResult
// 	err error
// }

// func (s *testTagger) Process(string, *api.SegmenterResult) (*api.TaggerResult, error) {
// 	return s.res, s.err
// }

// type testLex struct {
// 	res *api.SegmenterResult
// 	err error
// }

// func (s *testLex) Process(string) (*api.SegmenterResult, error) {
// 	return s.res, s.err
// }

// type testPreprocessor struct {
// 	err error
// }

// func (s *testPreprocessor) Process(text string) (string, error) {
// 	return text, s.err
// }

// func newTestData(t service.Tagger, s service.Segmenter) *service.Data {
// 	return &service.Data{Tagger: t, Segmenter: s, Preprocesor: &testPreprocessor{}}
// }
