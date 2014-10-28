package schemas

import (
	"fmt"
	"net/url"
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

	refUrl, _ := url.Parse(documentURL)

	mapDoc := doc.Document.(map[string]interface{})

	document := doc.Document

	if err == nil && refUrl.Fragment != "" {
		// If we have a fragment, grab it.
		document, _, err = resolvedURL.GetPointer().Get(document)
	}

	if err != nil {
		return nil, err
	}

	mapDoc = document.(map[string]interface{})

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

	spew.Dump(GetServiceMethods("/protocol/power"))

	//GetSchema("/model/device")
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
