# firecheck

A tool written in go to check permissions (R W D) on firebase instances.

## Installing

Requires [Go](https://golang.org/dl/)

`go get -u github.com/bp0lr/firecheck`

## How To Use:

Examples: 
- `firecheck -u "https://testdb-1ec08.firebaseio.com"`
- `firecheck -u "https://testdb-1ec08.firebaseio.com" -H "foo: bar" -o result.txt -s`
- `cat urls.txt | firecheck -o result.txt -s`

Options:
```
-H, --header stringArray   Add custom Headers to the request
-o, --output string        Output file to save the results to
-p, --proxy string         Add a HTTP proxy
-r, --random-agent         Set a random User Agent
-s, --simple               Display only the url without R W D
-u, --url string           The firebase url to test
-m, --user string          Add your username for write POC
-v, --verbose              Display extra info about what is going on
-w, --workers int          Workers amount (default 50)
```

## Practical Use

Use this tool in conbination with others for max results.

one linner using:
  - getJS (https://github.com/003random/getJS)
  - fget (https://github.com/bp0lr/fget)
  - gf (https://github.com/tomnomnom/gf)
  - js-beautify (npm js-beautify)
  - httpx (https://github.com/projectdiscovery/httpx)

```cat urls.txt | getJS --complete --resolve | fget -w 50 -r -f -o . && find results/ -iname '*.js' -exec bash -c "js-beautify --quiet -o {}.ok.js {} > /dev/null 2>&1" \; && find results/ -type f -name "*.js" \! -name "*.ok.js" -exec rm -f {} \; && for D in `find results/ -type d`; do for file in `find ${D} -type f`; do gf firebase_secrets ${file} | awk -F: '{print $3}'  >> gf.txt; done; done && cat gf.txt | httpx -silent | firecheck -v -o firebase.txt```
