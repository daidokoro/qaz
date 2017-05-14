package commands

type hooks struct {
	pre    []string `yaml:"pre,omitempty",json:"pre,omitempty"`
	post   []string `yaml:"post,omitempty",json:"post,omitempty"`
	update []string `yaml:"update,omitempty",json:"update,omitempty"`
}

// invoke lambda hook - stage pre/post
// func (h *hooks) invoke(stage string) error {
// 	switch stage {
// 	case "pre":
// 		for _, e := range h.pre {
// 			source := strings.Split(e, "@")
//
// 			if len(source) < 2 {
// 				return fmt.Errorf("invalid hook detected: [%s]", e)
// 			}
//
// 			f := function{
// 				payload: []byte(source[0]),
// 				name:    source[1],
// 			}
//
//       f.Invoke(sess)
// 		}
// 	case "post":
// 	default:
// 		//TODO: no default
// 	}
//
// }
