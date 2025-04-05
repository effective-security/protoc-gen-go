package enumgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_formatDisplayName(t *testing.T) {
	assert.Equal(t, "Test", FormatDisplayName("Test"))
	assert.Equal(t, "test", FormatDisplayName("test"))
	assert.Equal(t, "Test Data", FormatDisplayName("TestData"))
	assert.Equal(t, "AWS Name", FormatDisplayName("AWSName"))
	assert.Equal(t, "S3 Location", FormatDisplayName("S3Location"))
	assert.Equal(t, "EC2 Instance", FormatDisplayName("EC2Instance"))
}
