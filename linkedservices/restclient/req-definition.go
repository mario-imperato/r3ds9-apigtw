package restclient

const (
	HttpScheme    = "http"
	HttpLocalhost = "localhost"
)

/* type RequestOption func(o *RequestDefinition)

type RequestDefinition struct {
	Scheme   string          `mapstructure:"scheme"`
	Hostname string          `mapstructure:"hostname"`
	Port     int             `mapstructure:"port"`
	Path     string          `mapstructure:"path"`
	Method   string          `mapstructure:"method"`
	Headers  []NameValuePair `mapstructure:"headers"`
	Body     interface{}     `mapstructure:"body"`
}


func (req *RequestDefinition) Url() string {

	var sb strings.Builder

	if req.Scheme == "" {
		req.Scheme = HttpScheme
	}

	sb.WriteString(req.Scheme)
	sb.WriteString("://")

	if req.Hostname != "" {
		sb.WriteString(req.Hostname)
	} else {
		sb.WriteString(HttpLocalhost)
	}

	if req.Port != 0 {
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(req.Port))
	}

	if req.Path == "" {
		req.Path = "/"
	}

	sb.WriteString(req.Path)

	return sb.String()
}
*/
