/**
 * Copyright 2025 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	xchange "github.com/rdkcentral/xconfadmin/shared/change"
	xadmin_logupload "github.com/rdkcentral/xconfadmin/shared/logupload"

	ds "github.com/rdkcentral/xconfwebconfig/db"
	xwchange "github.com/rdkcentral/xconfwebconfig/shared/change"
	"github.com/rdkcentral/xconfwebconfig/shared/logupload"

	"github.com/rdkcentral/xconfwebconfig/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const telemetryJsonConfig = "{\n    \"Description\":\"Test Json Data\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":43200,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"RootName\":\"someNewRootName\",\n    \"Parameter\":\n        [\n            { \"type\": \"dataModel\", \"reference\": \"Profile.Name\"}, \n            { \"type\": \"dataModel\", \"reference\": \"Profile.Version\"},\n            { \"type\": \"grep\", \"marker\": \"Connie_marker1\", \"search\":\"restart 'lock to rescue CMTS retry' timer\", \"logFile\":\"cmconsole.log\" }\n\n        ],\n    \"HTTP\": {\n        \"URL\":\"https://test.net\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\",\n        \"RequestURIParameter\": [\n            {\"Name\":\"profileName\", \"Reference\":\"Profile.Name\" },\n            {\"Name\":\"reportVersion\", \"Reference\":\"Profile.Version\" }\n        ]\n\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n\n}"
const changedTelemetryJsonConfig = "{\n    \"Description\":\"Changed Name Json Data\",\n    \"Version\":\"0.1\",\n    \"Protocol\":\"HTTP\",\n    \"EncodingType\":\"JSON\",\n    \"ReportingInterval\":43200,\n    \"TimeReference\":\"0001-01-01T00:00:00Z\",\n    \"RootName\":\"someNewRootName\",\n    \"Parameter\":\n        [\n            { \"type\": \"dataModel\", \"reference\": \"Profile.Name\"}, \n            { \"type\": \"dataModel\", \"reference\": \"Profile.Version\"},\n            { \"type\": \"grep\", \"marker\": \"Connie_marker1\", \"search\":\"restart 'lock to rescue CMTS retry' timer\", \"logFile\":\"cmconsole.log\" }\n\n        ],\n    \"HTTP\": {\n        \"URL\":\"https://test.net\",\n        \"Compression\":\"None\",\n        \"Method\":\"POST\",\n        \"RequestURIParameter\": [\n            {\"Name\":\"profileName\", \"Reference\":\"Profile.Name\" },\n            {\"Name\":\"reportVersion\", \"Reference\":\"Profile.Version\" }\n        ]\n\n    },\n    \"JSONEncoding\": {\n        \"ReportFormat\":\"NameValuePair\",\n        \"ReportTimestamp\": \"None\"\n    }\n\n}"

func TestTelemetryTwoProfileCreateHandler(t *testing.T) {

	DeleteAllEntities()

	p := createTelemetryTwoProfile()

	entryByte, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(entryByte))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	createdProfile := unmarshalTelemetryTwoProfile(rr.Body.Bytes())

	assert.Equal(t, p, createdProfile)

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Equal(t, *p, *dbProfile, "profile to create should match created profile in database")
}

func TestTelemetryTwoProfileCreateChangeHandlerAndApproveIt(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryTwoProfile()

	requestStr, _ := json.Marshal(p)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile/change?%v", queryParams)

	r := httptest.NewRequest("POST", url, bytes.NewReader(requestStr))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusCreated, rr.Code)

	change := unmarshalChangeTwo(rr.Body.Bytes())

	assert.Empty(t, change.OldEntity, "old entity in create change should be nil")
	assert.Equal(t, *p, *change.NewEntity, "new entity should match profile to create")

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Empty(t, dbProfile, "profile before approval should not be present in database")

	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Equal(t, *p, *dbProfile, "profile to create should match created profile in database")

	approvedChange := xchange.GetOneApprovedTelemetryTwoChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved profile change should be created")
	assert.Empty(t, approvedChange.OldEntity, "old entity should not present")
	assert.Equal(t, *p, *approvedChange.NewEntity, "old entity should not present")
}

func TestTelemetryTwoProfileUpdateHandler(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, p.ID, p)

	changedProfile, _ := p.Clone()
	changedProfile.Jsonconfig = changedTelemetryJsonConfig

	requestStr, _ := json.Marshal(changedProfile)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(requestStr))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	updatedProfile := unmarshalTelemetryTwoProfile(rr.Body.Bytes())

	assert.Equal(t, *changedProfile, *updatedProfile)

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.NotEqual(t, p.Jsonconfig, dbProfile.Jsonconfig, "updated profile data should match")
	assert.Equal(t, updatedProfile.Jsonconfig, dbProfile.Jsonconfig, "updated profile data should match")
}

func TestTelemetryTwoProfileUpdateChangeHandler(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, p.ID, p)

	changedProfile, _ := p.Clone()
	changedProfile.Jsonconfig = changedTelemetryJsonConfig

	requestStr, _ := json.Marshal(changedProfile)
	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile/change?%v", queryParams)

	r := httptest.NewRequest("PUT", url, bytes.NewReader(requestStr))
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChangeTwo(rr.Body.Bytes())

	assert.Equal(t, *p, *change.OldEntity, "old entity should correspond to the profile before updating")
	assert.Equal(t, *changedProfile, *change.NewEntity, "new entity should correspond to the profile to create")

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Equal(t, *p, *dbProfile, "profile before approval should not be changed")

	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	dbProfile = logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Equal(t, *changedProfile, *dbProfile, "profile to create should match created profile in database")
	assert.Equal(t, changedTelemetryJsonConfig, dbProfile.Jsonconfig, "profile to create should match created profile in database")

	approvedChange := xchange.GetOneApprovedTelemetryTwoChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved profile change should be created")
	assert.Equal(t, *p, *approvedChange.OldEntity, "old entity should correspond to the profile before updating it")
	assert.Equal(t, *changedProfile, *approvedChange.NewEntity, "new entity should correspond to the changed profile")
}

func TestTelemetryTwoProfileDeleteHandler(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	ds.GetCachedSimpleDao().RefreshAll(ds.TABLE_TELEMETRY_TWO_PROFILES)

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Empty(t, dbProfile, "profile after removal should not exist in db")
}

func TestTelemetryTwoProfileDeleteChangeHandler(t *testing.T) {
	DeleteAllEntities()

	p := createTelemetryTwoProfile()
	ds.GetCachedSimpleDao().SetOne(ds.TABLE_TELEMETRY_TWO_PROFILES, p.ID, p)

	queryParams, _ := util.GetURLQueryParameterString([][]string{
		{"applicationType", "stb"},
	})
	url := fmt.Sprintf("/xconfAdminService/telemetry/v2/profile/change/%v?%v", p.ID, queryParams)

	r := httptest.NewRequest("DELETE", url, nil)
	rr := ExecuteRequest(r, router)
	assert.Equal(t, http.StatusOK, rr.Code)

	change := unmarshalChangeTwo(rr.Body.Bytes())

	assert.Equal(t, *p, *change.OldEntity, "old entity should correspond profile before removing")
	assert.Empty(t, change.NewEntity, "new entity should be empty")

	dbProfile := logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Equal(t, *p, *dbProfile, "profile before approval should not be changed")

	url = fmt.Sprintf("/xconfAdminService/telemetry/v2/change/approve/%v?%v", change.ID, queryParams)

	r = httptest.NewRequest("GET", url, nil)
	rr = ExecuteRequest(r, router)

	assert.Equal(t, http.StatusOK, rr.Code)

	ds.GetCachedSimpleDao().RefreshAll(ds.TABLE_TELEMETRY_TWO_PROFILES)

	dbProfile = logupload.GetOneTelemetryTwoProfile(p.ID)
	assert.Empty(t, dbProfile, "profile after approval should be removed")

	approvedChange := xchange.GetOneApprovedTelemetryTwoChange(change.ID)
	assert.NotEmpty(t, approvedChange, "approved profile change should be created")
	assert.Equal(t, *p, *approvedChange.OldEntity, "old entity should correspond to the profile before removing it")
	assert.Empty(t, approvedChange.NewEntity, "new entity should not be created")
}

func unmarshalTelemetryTwoProfile(b []byte) *logupload.TelemetryTwoProfile {
	var profile logupload.TelemetryTwoProfile
	err := json.Unmarshal(b, &profile)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return &profile
}

func unmarshalChangeTwo(b []byte) xwchange.TelemetryTwoChange {
	var change xwchange.TelemetryTwoChange
	err := json.Unmarshal(b, &change)
	if err != nil {
		panic(fmt.Errorf("error unmarshaling telemetry profile change"))
	}
	return change
}

func createTelemetryTwoProfile() *logupload.TelemetryTwoProfile {
	p := xadmin_logupload.NewEmptyTelemetryTwoProfile()
	p.ID = uuid.New().String()
	p.Name = "Test Telemetry 2 Profile"
	p.Jsonconfig = telemetryJsonConfig
	return p
}
