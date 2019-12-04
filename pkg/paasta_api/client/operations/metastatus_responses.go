// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/Yelp/paasta-tools-go/pkg/paasta_api/models"
)

// MetastatusReader is a Reader for the Metastatus structure.
type MetastatusReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *MetastatusReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewMetastatusOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewMetastatusInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewMetastatusOK creates a MetastatusOK with default headers values
func NewMetastatusOK() *MetastatusOK {
	return &MetastatusOK{}
}

/*MetastatusOK handles this case with default header values.

Detailed metastatus
*/
type MetastatusOK struct {
	Payload *models.MetaStatus
}

func (o *MetastatusOK) Error() string {
	return fmt.Sprintf("[GET /metastatus][%d] metastatusOK  %+v", 200, o.Payload)
}

func (o *MetastatusOK) GetPayload() *models.MetaStatus {
	return o.Payload
}

func (o *MetastatusOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.MetaStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewMetastatusInternalServerError creates a MetastatusInternalServerError with default headers values
func NewMetastatusInternalServerError() *MetastatusInternalServerError {
	return &MetastatusInternalServerError{}
}

/*MetastatusInternalServerError handles this case with default header values.

Metastatus failure
*/
type MetastatusInternalServerError struct {
}

func (o *MetastatusInternalServerError) Error() string {
	return fmt.Sprintf("[GET /metastatus][%d] metastatusInternalServerError ", 500)
}

func (o *MetastatusInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
