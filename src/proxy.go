package main 

import (
    "net/http"
    "io/ioutil"
    "encoding/xml"
    "strings"
    "bytes"
    "log"
)

type Process struct {
	XMLName	xml.Name `xml:"proc"`
	Exec	string `xml:"exec,attr"`
    Rss		string `xml:"rss,attr"`
}

type Command struct {
	Name string
	Size string
}

type Answer struct {
	Command []Command
}

type Root struct {
	Answer []Answer
}

func answer() *Answer {
	answer := new(Answer)
	response, err := http.Get("http://10.7.1.13:3000")
    if err != nil {
		log.Print(err)
	} else {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Print(err)
		} else {
		    decoder := xml.NewDecoder(bytes.NewReader(content))
		    XML: for {
				token, err := decoder.Token()
				switch {
					case err != nil && err.Error() != "EOF":
						log.Print(err)
					case token == nil:
						break XML
					default : {
						switch element := token.(type) {
							case xml.StartElement:
								if element.Name.Local == "proc" {
									var process Process
									decoder.DecodeElement(&process, &element)
									if strings.Contains(process.Exec, "log") {
										command := Command{Name:process.Exec,Size:process.Rss}
										log.Print(command)
										answer.Command = append(answer.Command, command)
									}
								}
						}
					}	
				}
			}
	    }
	}
	return answer
}

func root() *Root {
	root := new(Root)
	answers := make(chan *Answer)
	for i := 0; i < 3; i++ {
		go func(){
			answers <- answer()
		}()
	}
	for i := 0; i < 3; i++ {
		answer := <- answers
		root.Answer = append(root.Answer, *answer)
	}
	return root
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	enc := xml.NewEncoder(w)
	enc.Indent("","  ")
    err := enc.Encode(root())
    if err != nil {
       http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
	http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
