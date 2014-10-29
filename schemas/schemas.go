package schemas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/gojsonschema"
	"github.com/xeipuuv/gojsonreference"
)

var log = logger.GetLogger("schemas")

var root = "http://schema.ninjablocks.com/"
var rootURL, _ = url.Parse(root)
var filePrefix = config.MustString("installDirectory") + "/sphere-schemas/"
var fileSuffix = ".json"

var schemaPool = gojsonschema.NewSchemaPool()

func init() {
	schemaPool.FilePrefix = &filePrefix
	schemaPool.FileSuffix = &fileSuffix
}

func Validate(schema string, obj interface{}) (*string, error) {

	log.Debugf("schema-validator: validating %s %v", schema, obj)

	doc, err := GetSchema(schema)

	if err != nil {
		return nil, fmt.Errorf("Failed to get document: %s", err)
	}

	// Try to validate the Json against the schema
	result := doc.Validate(obj)

	messages := ""

	// Deal with result
	if !result.Valid() {
		// Loop through errors
		for _, desc := range result.Errors() {
			messages += fmt.Sprintf("%s\n", desc)
		}
	}

	return &messages, nil
}

func GetServiceMethods(service string) ([]string, error) {
	doc, err := GetDocument(service+"#/methods", true)

	if err != nil && fmt.Sprintf("%s", err) != "Object has no key 'methods'" {
		return nil, err
	}

	methods := make([]string, 0, len(doc))
	for method := range doc {
		methods = append(methods, method)
	}

	return methods, nil
}

type flatItem struct {
	path  []string
	value interface{}
}

func flatten(input interface{}, lpath []string, flattened []flatItem) []flatItem {
	if lpath == nil {
		lpath = []string{}
	}
	if flattened == nil {
		flattened = []flatItem{}
	}

	if reflect.ValueOf(input).Kind() == reflect.Map {
		for rkey, value := range input.(map[string]interface{}) {
			flattened = flatten(value, append(lpath, rkey), flattened)
		}
	} else {
		flattened = append(flattened, flatItem{lpath, input})
	}

	return flattened
}

type TimeSeriesDatapoint struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

/*
* GetEventTimeSeriesData converts an event payload to 0..n time series data points.
* NOTE: The payload must already have been validated. No validation is done here.
* NOTE: This accepts the json payload. So either a simple type or map[string]interface{}
*
* @param value {interface{}} The payload of the event. Can be null if there is no payload
* @param eventSchemaUri {string} The URI of the schema defining the event (usually ends with #/events/{name})
* @returns {Array} An array of records that need to be saved to a time series db
 */
func GetEventTimeSeriesData(value interface{}, serviceSchemaUri, event string) ([]TimeSeriesDatapoint, error) {

	// We don't want a pointer, just grab the actual value
	if reflect.ValueOf(value).Kind() == reflect.Ptr {
		value = reflect.ValueOf(value).Elem().Interface()
	}

	var timeseriesData = make([]TimeSeriesDatapoint, 0)

	eventSchema, err := GetDocument(serviceSchemaUri+"#/events/"+event, true)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve event schema: %s event: %s error: %s", serviceSchemaUri, event, err)
	}

	log.Debugf("Finding time series data for service: %s event: %s from payload: %v", serviceSchemaUri, event, value)

	if _, ok := eventSchema["value"]; ok {
		// The event emits a value.
		flat := flatten(value, nil, nil)

		for _, point := range flat {
			log.Debugf("-- Checking: %v", point)

			refPath := fmt.Sprintf("#/events/%s/value", event)
			if len(point.path) > 0 {
				// Not the root value
				refPath = strings.Join(append([]string{refPath}, point.path...), "/properties/")
			}
			log.Debugf("Created path %s", refPath)

			pointSchema, err := GetDocument(serviceSchemaUri+refPath, true)

			if err != nil {
				// As the data has been validated, this *shouldn't* happen. BUT we might be allowing unknown properties through.
				log.Warningf("Unknown property %s in service %s event %s. error: %s", refPath, serviceSchemaUri, event, err)
			} else {

				if timeseriesType, ok := pointSchema["timeseries"]; ok {

					dp := TimeSeriesDatapoint{
						Path: strings.Join(point.path, "."),
						Type: timeseriesType.(string),
					}

					if timeseriesType == "value" || timeseriesType == "boolean" {
						dp.Value = point.value
					}

					// The only other type is 'event', which doesn't have or need a value

					timeseriesData = append(timeseriesData, dp)
				}
			}
		}
	}

	return timeseriesData, nil
}

func GetDocument(documentURL string, resolveRefs bool) (map[string]interface{}, error) {
	resolvedURL, err := resolveUrl(rootURL, documentURL)
	if err != nil {
		return nil, err
	}

	localURL := useLocalUrl(resolvedURL)

	doc, err := schemaPool.GetDocument(localURL)
	if err != nil {
		return nil, err
	}

	refURL, _ := url.Parse(documentURL)

	document := doc.Document

	if err == nil && refURL.Fragment != "" {
		// If we have a fragment, grab it.
		document, _, err = resolvedURL.GetPointer().Get(document)
	}

	if err != nil {
		return nil, err
	}

	mapDoc := document.(map[string]interface{})

	if resolveRefs {
		if ref, ok := mapDoc["$ref"]; ok && ref != "" {
			log.Debugf("Got $ref: %s", ref)
			var resolvedRef, err = resolveUrl(resolvedURL.GetUrl(), ref.(string))
			log.Debugf("resolved %s to %s", ref.(string), resolvedRef.GetUrl().String())
			if err != nil {
				return nil, err
			}
			return GetDocument(resolvedRef.String(), true)
		}
	}

	return mapDoc, nil
}

var schemasCache = make(map[string]*gojsonschema.JsonSchemaDocument)

func GetSchema(documentURL string) (*gojsonschema.JsonSchemaDocument, error) {

	resolved, err := resolveUrl(rootURL, documentURL)
	if err != nil {
		return nil, err
	}
	localRef := useLocalUrl(resolved)
	local := localRef.GetUrl().String()

	schema, ok := schemasCache[local]
	if !ok {
		log.Debugf("Cache miss on '%s'", resolved.GetUrl().String())
		schema, err = gojsonschema.NewJsonSchemaDocument(local, schemaPool)
		schemasCache[local] = schema
	}
	return schema, err
}

func useLocalUrl(ref gojsonreference.JsonReference) gojsonreference.JsonReference {
	// Grab ninjablocks schemas locally

	local := strings.Replace(ref.GetUrl().String(), root, "file:///", 1)
	log.Infof("Fetching document from %s", local)
	localURL, _ := gojsonreference.NewJsonReference(local)
	return localURL
}

func resolveUrl(root *url.URL, documentURL string) (gojsonreference.JsonReference, error) {
	ref, err := gojsonreference.NewJsonReference(documentURL)
	if err != nil {
		return ref, err
	}
	resolvedURL := root.ResolveReference(ref.GetUrl())

	return gojsonreference.NewJsonReference(resolvedURL.String())
}

func main() {
	//spew.Dump(Validate("/protocol/humidity#/events/state/value", "hello"))
	//spew.Dump(Validate("protocol/humidity#/events/state/value", 10))

	// TODO: FAIL! min/max not taken care of!
	//spew.Dump(Validate("/protocol/humidity#/events/state/value", -10))

	//spew.Dump(GetServiceMethods("/protocol/power"))
	/*	doc, _ := GetDocument("/protocol/humidity", true)
		flattened := flatten(doc, []string{}, make([]flatItem, 0))
		spew.Dump(flattened)*/

	spew.Dump(GetEventTimeSeriesData(10, "/protocol/humidity", "state"))

	var payload = &testVal{
		Rumbling: true,
		X:        0.5,
		Y:        -0.1,
		Z: &testValSize{
			Hello:   10,
			Goodbye: 20,
		},
	}

	jsonBytes, _ := json.Marshal(payload)
	var jsonPayload interface{}
	_ = json.Unmarshal(jsonBytes, &jsonPayload)

	points, _ := GetEventTimeSeriesData(&jsonPayload, "/protocol/game-controller/joystick", "state")

	js, _ := json.Marshal(points)

	log.Infof("Points: %s", js)

	//spew.Dump(GetEventTimeSeriesData(nil, "/protocol/humidity", "state"))
}

type testVal struct {
	Rumbling bool         `json:"rumbling"`
	X        float64      `json:"x"`
	Y        float64      `json:"y"`
	Z        *testValSize `json:"z"`
}

type testValSize struct {
	Hello   int `json:"hello"`
	Goodbye int `json:"goodbye"`
}
