# APIDoc
> Current version: beta-0.3.3

<br>
<p align="center">
    <img alt="apidoc" src="https://github.com/spaceavocado/apidoc/raw/master/assets/apidoc.png" width="240">
</p>

<p align="center">
  Generate RESTful API documentation from GO source files into the <a href="https://swagger.io/specification/">OpenAPI v3.0.2</a> specification (formal Swagger 2.0 Specification).
</p>

<p align="center">
<a href="https://travis-ci.org/spaceavocado/apidoc.svg?branch=master"><img alt="Travis Status" src="https://travis-ci.org/spaceavocado/apidoc.svg?branch=master"></a> <a href="https://codecov.io/gh/spaceavocado/apidoc">
  <img src="https://codecov.io/gh/spaceavocado/apidoc/branch/master/graph/badge.svg" />
</a> <a href="https://goreportcard.com/badge/github.com/spaceavocado/apidoc"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/spaceavocado/apidoc"></a> <a href="https://www.codacy.com/app/davidhorak/apidoc?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=spaceavocado/apidoc&amp;utm_campaign=Badge_Grade"><img src="https://api.codacy.com/project/badge/Grade/ed701d5e42f3426d832de973906ba19b"/></a> <a href="https://app.fossa.io/projects/git%2Bgithub.com%2Fspaceavocado%2Fapidoc?ref=badge_shield" alt="FOSSA Status"><img src="https://app.fossa.io/api/projects/git%2Bgithub.com%2Fspaceavocado%2Fapidoc.svg?type=shield"/></a> <a href="https://godoc.org/github.com/spaceavocado/apidoc"><img alt="Go Doc" src="https://godoc.org/github.com/spaceavocado/apidoc?status.svg"></a>
</p>

## Table of Contents
- [APIDoc](#apidoc)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Examples](#examples)
  - [Getting Started](#getting-started)
- [API Annotation in Comments](#api-annotation-in-comments)
  - [Main Section](#main-section)
    - [Supported Tags](#supported-tags)
    - [Example](#example)
  - [An Endpoint](#an-endpoint)
    - [Supported Tags](#supported-tags-1)
    - [Param Tag](#param-tag)
      - [name](#name)
      - [in](#in)
      - [type](#type)
    - [Body Tag](#body-tag)
      - [reference](#reference)
    - [Wrapper Tag](#wrapper-tag)
      - [Example](#example-1)
    - [Response Tag](#response-tag)
      - [code](#code)
      - [type](#type-1)
      - [reference](#reference-1)
    - [Path Tag](#path-tag)
      - [path](#path)
      - [method](#method)
    - [Example Endpoint Annotation](#example-endpoint-annotation)
  - [gorilla/mux Handler Functions](#gorillamux-handler-functions)
    - [Notes](#notes)
  - [gorilla/mux Subrouter](#gorillamux-subrouter)
    - [Subrouter Annotation](#subrouter-annotation)
      - [Example](#example-2)
  - [Mime Types Annotation](#mime-types-annotation)
  - [Struct Annotation](#struct-annotation)
  - [Data Types Conversion](#data-types-conversion)
- [Tips](#tips)
  - [Annotation over Multiple Lines](#annotation-over-multiple-lines)
  - [Array References](#array-references)
  - [And Endpoint With Many Decralarions](#and-endpoint-with-many-decralarions)
    - [Example](#example-3)
- [APIDoc CLI](#apidoc-cli)
- [About the Project](#about-the-project)
- [Contributing](#contributing)
  - [Pull Request Process](#pull-request-process)
- [License](#license)

## Summary
APIDoc extracts the API documentation annotation from your GO source files, recursively resoles struct references, and it generates the YAML [OpenAPI v3.0.2](https://swagger.io/specification/) spec. file, which could be tested in the [Swagger Editor](https://editor.swagger.io/) and quickly integrated with [Swagger UI](https://swagger.io/tools/swagger-ui/download/).

The generator is also able to read [gorilla/mux](https://github.com/gorilla/mux) **Handler** and **HandlerFunc** func signature to automatically generate the `@router` tag, and `@param` tag/s. [See gorilla/mux Handler Functions](#gorillamux-handler-functions).

> Note: It does not generate/support all OpenAPI v3.0.2 blocks, the tool handles just a subset required by our internal needs, it might be extended if there will be high demand on extension, or if anyone would like to contribute.

## Examples
[Basic API Project](https://github.com/spaceavocado/apidoc/tree/master/example)

## Getting Started
1. Add annotation into your API source code files, [See API Annotation in Comments](#api-annotation-in-comments).
   
2. Get the APIDoc by using:
    ```sh
    go get -u github.com/spaceavocado/apidoc
    ```
    > This will create a binary in your **$GOPATH/bin** folder called `apidoc` (Mac/Unix) or `apidoc.exe` (Windows).

3. Now you can run `apidoc` command from your shell.
   > Make sure that **$GOPATH/bin** is in your Environment Variables:
    - **Linux**:
      ```sh
      export PATH=$PATH:$GOPATH/bin
      ```
      *Note*: assumption that your have correctly set $GOPATH env. variable.

    - **Windows**: "Control Panel" > "System" > "Edit the system environment variables" > "Advanced" > "Environment Variables" > "Path" > "Edit". and add the directory.

4. Run `apidoc` in the your project's root folder. This will extract and process your annotation and it generates the output YAML file.
    ```sh
    apidoc -m main.go -e handler -o docs/api
    ```
    *Note*: the example shows the default flag values, for more details, [See APIDoc CLI](#apidoc-cli).

5. Preview the documentation in the [Swagger Editor](https://editor.swagger.io/), i.e. put the openapi.yaml content into the editor.

# API Annotation in Comments
An API annotation is represented as an code comment starting with **@** symbol followed by documentation tag and it value or parameters. Examples:

* `// @title Example RESTful API`, tag **title** with value "Example RESTful API"
* `// @server api.domain.com/v3 Production`, tag **server** with params: "api.domain.com/v3", "Production" 

The the value of the tag or the last parameter of the tag captures the multiline comments, example:
```go
// @desc Lorem ipsum dolor sit amet, consectetur adipiscing
// elit. Nullam rhoncus magna nunc, in faucibus metus pulvinar
// et. Mauris pellentesque 
```
> The value of desc tag is captured as: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam rhoncus magna nunc, in faucibus metus pulvinar et. Mauris pellentesque"
```go
// @server api.domain.com/v3 Lorem ipsum dolor sit amet, consectetur
// elit. Nullam rhoncus magna nunc.
```
> The last parameter, **description** of the server tag is captured as: "Lorem ipsum dolor sit amet, consectetur elit. Nullam rhoncus magna nunc."
  
## Main Section
This section should be located in the **main** file, passed to the [APIDoc CLI](#apidoc-cli) in **-m** flag (defaults to **main.go**).

### Supported Tags
> Note: **()** within **Annotation** indicates an annotation parameter captured by the generator.

| Annotation                 | Description                                                                                             | OpenAPI Spec.                                                      | Example                                                      |
| -------------------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ | ------------------------------------------------------------ |
| title (value)              | **REQUIRED**. The title of the application.                                                             | https://swagger.io/specification/#infoObject field: title          | // @title Example RESTful API                                |
| desc (value)               | A short description of the application.                                                                 | https://swagger.io/specification/#infoObject field: description    | // @desc An example of publicly accessible API               |
| ver (value)                | **REQUIRED**. The version of the OpenAPI document                                                       | https://swagger.io/specification/#infoObject field: version        | // @ver 1.0.0                                                |
| terms (value)              | A URL to the Terms of Service for the API.                                                              | https://swagger.io/specification/#infoObject field: termsOfService | // @terms https://domain.com/api/terms                       |
| contact.name (value)       | The identifying name of the contact person/organization.                                                | https://swagger.io/specification/#contactObject field: name        | // @contact.name API Support                                 |
| contact.url (value)        | The URL pointing to the contact information.                                                            | https://swagger.io/specification/#contactObject field: url         | // @contact.url https://domain.com/api/support               |
| contact.email (value)      | The email address of the contact person/organization.                                                   | https://swagger.io/specification/#contactObject field: email       | // @contact.email support@domain.com                         |
| lic.name (value)           | The license name used for the API.                                                                      | https://swagger.io/specification/#licenseObject field: name        | // @lic.name Apache 2.0                                      |
| lic.url (value)            | A URL to the license used for the API.                                                                  | https://swagger.io/specification/#licenseObject field: url         | // @lic.url https://www.apache.org/licenses/LICENSE-2.0.html |
| server (url) (description) | An object representing a Server <br><br>Note: There might be many @server tags within the main section. | https://swagger.io/specification/#serverObject                     | // @server api.domain.com/v3 Production                      |

### Example
```go
// @title An example authentication API
// @desc Publicly accessible authentication REST API.
// @terms https://domain.com/docs/api/terms
//
// @contact.name API Support
// @contact.url https://domain.com/contact
// @contact.email support@domain.com
//
// @lic.name Apache 2.0
// @lic.url https://www.apache.org/licenses/LICENSE-2.0.html
//
// @ver 1.0
// @server https://auth.domain.com/v3 Production API
// @server https://auth.dev.domain.com/v3 Development API
func main() {}
```

## An Endpoint
An endpoint is being considered as a API comment annotation block found within any file located inside the **endpoints** root folder, passed to the [APIDoc CLI](#apidoc-cli) in **-e** flag (defaults to **./**).
> For better performance is highly recommended to pass the endpoints root folder as a flag to the CLI to avoid unnecessary file processing.

### Supported Tags
> Note: **()** within **Annotation** indicates an annotation parameter captured by the generator.

| Annotation                                                 | Description                                                                                                                                                                                   | OpenAPI Spec.                                                        | Example                                                                                               |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| summary (value)                                            | A short summary of what the operation does.                                                                                                                                                   | https://swagger.io/specification/#operationObject field: summary     | @sumamry Lorem ipsum dolor                                                                            |
| desc (value)                                               | A verbose explanation of the operation behavior.                                                                                                                                              | https://swagger.io/specification/#operationObject field: description | @desc Lorem ipsum dolor sit amet, consectetur adipiscing eli                                          |
| id (value)                                                 | Unique string used to identify the operation. The id MUST be unique among all operations described in the API. The operationId value is case-sensitive.                                       | https://swagger.io/specification/#operationObject field: operationId | // @id login                                                                                          |
| tag (array)                                                | A list of tags for API documentation control. Tags can be used for logical grouping of operations by resources or any other qualifier.<br><br>Note: Comma separated value, or a single value. | https://swagger.io/specification/#operationObject field: tags        | // @tag User<br><br>// @tag User, Authentication                                                      |
| accept (array)                                             | A list of accepted request content mime types<br><br>[See Mime Types Annotation](#mime-types-annotation)                                                                                      | n/a                                                                  | // @accept json<br><br>// @accept application/json, application/x-www-form-urlencoded                 |
| produce (array)                                            | **REQUIRED**. A list of supported content response mime types<br><br>[See Mime Types Annotation](#mime-types-annotation)                                                                      | n/a                                                                  | // @produce json<br><br>// @produce application/json, application/x-www-form-urlencoded               |
| param (name) (in) {(type)} (required) (description)        | Describes a single operation parameter. <br><br>[See Param Tag](#param-tag)<br><br>Note: There might be many @param tags within the endpoint block.                                           | https://swagger.io/specification/#parameterObject                    | // @param token path {string} true Security Token                                                     |
| body (reference)                                           | Describes a single request body.<br><br>[See Body Tag](#body-tag)                                                                                                                             | https://swagger.io/specification/#requestBodyObject                  | //@body model.Login                                                                                   |
| swrap (reference) (field pointer)                          | Success object wrapper. If this tag is set, any success response object is being wrapped with this object on the desired **pointer field**<br><br>[See Wrapper Tag](#wrapper-tag)             | n/a                                                                  | // @swrap response.Success                                                                            |
| success (code) {(type)} (reference or empty) (description) | **REQUIRED**. Describes a single success response from an API Operation.<br><br>[See Response Tag](#response-tag)                                                                             | https://swagger.io/specification/#responseObject                     | // @success 200 {object} response.Success OK<br><br>// @success 200 {string} OK                       |
| fwrap (reference) (field pointer)                          | Failure object wrapper. If this tag is set, any failure response object is being wrapped with this object on the desired **pointer field**<br><br>[See Wrapper Tag](#wrapper-tag)             | n/a                                                                  | // @fwrap response.Error                                                                              |
| failure (code) {(type)} (reference or empty) (description) | Describes a single failure response from an API Operation.<br><br>[See Response Tag](#response-tag)                                                                                           | https://swagger.io/specification/#responseObject                     | // @failure 401 {object} response.AuthError Unauthorized<br><br>// @failure 401 {string} Unauthorized |
| subrouter (value)                                          | Name of the subrouter used for this endpoint. <br><br>[See gorilla/mux Subrouter](#gorillamux-subrouter)                                                                                      | n/a                                                                  | // @subrouter user [post]                                                                             |
| router (path) [(method)]                                   | **REQUIRED**. Describes the operations available on a single path, i.e. endpoint URL<br><br>[See Path Tag](#path-tag)                                                                                       | https://swagger.io/specification/#pathItemObject                     | // @router /login [post]                                                                              |

### Param Tag
> *Annotation:* param (name) (in) {(type)} (required) (description)

This section contains detailed information about the **param** tag
#### name 
* The name of the parameter. Parameter names are case sensitive.
* If `in` is "path", the name field MUST correspond to the associated router url parameter, e.g.:
   `// @param token path {string} true Security token` must match with `// @router /auth/{token} [get]`

#### in 
* The location of the parameter. Possible values are "query", "header", "path" or "cookie".

#### type
* the outer "{}" are used just as visual separators of type field, i.e. their are not required.
  E.g. `// @param token path {string} true Security token` = `// @param token path string true Security token`
* Param tag does not support struct type, just base go types, [See Type Mapping](#type-mapping)
* **Array annotation**: `{[]string}`, `{[]int}`, etc.

### Body Tag
> *Annotation:* body (reference)

#### reference
* go structure used as request model, example: `// @body model.Login`
  
  ```go
  // Login struct located in "model" package,
  // referenced in the go file where the endpoint
  // annotation is set
  type Login struct {
    // User's email address
    Username string `json:"username"`
    Password string
  }
  ```
  will be resolved as:
  ```yaml
  schema:
    type: object
    parameters:
      username:
        type: string
        description: User's email address
      password:
        type: string
  ```
* The reference structure is being resolved recursively, i.e. it might contain fields referencing other go struct.
* [See Struct Annotation](#struct-annotation) for more details.

### Wrapper Tag
> *Annotation:* swrap (reference) (pointer field), fwrap (reference) (field pointer)

If this tag is set, any success/failure response object is being wrapped with this object on the desired pointer field. This is useful if you are using a generic API response object with various data field.

#### Example
```go
type APISuccess struct {
  Status string `json:"status"`
  Data interface{} `json:"data"`
}
type Profile struct {
  Firstname int `json:"firstname"`
  Lastname string `json:"lastname"`
}
```

`// @swrap APISuccess data` and `// @success 200 {object} Profile OK` will produce:
```yaml
"200":
  description: OK
  content:
    application/json:
      schema:
        type: object
        properties:
          status:
            type: string
          data:
            schema:
              $ref: "#/components/schemas/Profile"
```

* The reference structure is being resolved recursively, i.e. it might contain fields referencing other go struct.
* [See Struct Annotation](#struct-annotation) for more details.

### Response Tag
> *Annotation:* success (code) {(type)} (reference or string) (description), failure (code) {(type)} (reference or string) (description)

#### code
* Any valid HTTP reponse code

#### type
* The outer "{}" are used just as visual separators of type field, i.e. their are not required.
  E.g. `// @success 200 {string} string OK` = `// @success 200 string string OK`
* Supported: object, string
  * if set to object, the next parameter expected is **reference**
  * if set to string, next parameter is skipped (reference)

#### reference
* Reference response go struct
* **Array annotation**: `{[]response.Car}`, etc.
* The reference structure is being resolved recursively, i.e. it might contain fields referencing other go struct.
* [See Struct Annotation](#struct-annotation) for more details.

### Path Tag
> *Annotation:* router (path) [(method)]

#### path 
* It might contain the "path" parameters, it this format: `/login/{token}` where token is the name of param defined in the [Param Tag](#patam-tag).

#### method
* HTTP method
* The outer "[]" are used just as visual separators of method field, i.e. their are not required.
  E.g. `// @router /login [post]` = `// @router /login post`
* It might contain an array of methods, e.g.: `[post, put]`

### Example Endpoint Annotation
```go
// Login request
// @summary Login user
// @desc Authentication request
// @id login
// @tag Authentication
// @accept json, application/x-www-form-urlencoded
// @produce json
// @body model.Login
// @swrap response.Data data
// @success 200 {object} request.Token OK
// @failure 401 {object} response.Error Unauthorized.
// See error code and error message for more details.
// @failure 500 {string} Internal Server Error
// @router /login [post]
func handler(w http.ResponseWriter, r *http.Request) {}

// @summary Activate user
// @desc User activation via the token request.
// @id activate
// @tag Registration
// @param token path {string} true Activation token
// @produce json
// @success 200 {string} OK
// @failure 500 {string} Internal Server Error
// @router /registration/activate/{token} [get]
func handler(w http.ResponseWriter, r *http.Request) {}

// Handlers with passed gorilla mux router
//
// In this example the @router tag is being resolved
// from the HandleFunc or Handle gorilla *mux.Router func.
// @param tags ara also resolved from the func path string.
//
// These tags will be extracted automatically:
// @router /person/{id} [get]
// @param id path {string} true
func Handlers(r *mux.Router) {
  // GetPerson handler
  // @summary Person
  // @desc Get person by ID.
  // @id person
  // @tag Person
  // @produce json
  // @success 200 {object} Person OK
  // @failure 500 {string} Internal Server Error
  r.HandleFunc("/person/{id:[0-9]+}", GetPerson).Methods("GET")
}
```

## gorilla/mux Handler Functions
`@router` tag and  `@param` tags could automatically resolved if the endpoint annotation is placed above the [gorilla/mux](https://github.com/gorilla/mux) Handler or HandlerFunc function:
```go
// GetPerson handler
// @summary Person
// @desc Get person by ID.
// @id person
// @tag Person
// @produce json
// @success 200 {object} Person OK
// @failure 500 {string} Internal Server Error
r.HandleFunc("/person/{id:[0-9]+}", GetPerson).Methods("GET")
```
Automatically resolved tags will be be:
* `@router /person/{id} [get]`
* `@param id path {string} true`

### Notes
* This process is skipped if the endpoint annotation contains the `@router` tag
* If the endpoint annotation contains a `@param` tag found in the func path parameter, it is not being overwritten by the this process nor double annotated.
* Please [see gorilla/mux Subrouter](#gorillamux-subrouter) section for more information how to work with gorilla/mux subrouters.

## gorilla/mux Subrouter
An endpoint can use `subrouter` tag to connect a endpoint with the subrouter to resolve the final endpoint URL.

### Subrouter Annotation
| Annotation       | Description                   | Example             |
| ---------------- | ----------------------------- | ------------------- |
| router (name)    | Name of the subrouter.        | // @router account  |
| subrouter (name) | Name of the parent subrouter. | // @subrouter admin |

The annotation must be placed above the **gorilla.mux Subrouter** method anywhere within the Main file or in any file within the endpoints root folder.

#### Example
```go
// @router admin
r.PathPrefix("/admin").Subrouter()

// @router user
// @subrouter admin
r.PathPrefix("/user").Subrouter()

// @summary List of users
// @produce json
// @success 200 {object} Person OK
// @subrouter user
r.HandleFunc("/list", GetPersonList).Methods("GET")
```
> The URL resolved for the "List of users" endpoint will be `/admin/user/list`

## Mime Types Annotation
| Mime Type                         | Annotation                              |
| --------------------------------- | --------------------------------------- |
| text/plain                        | text, text/plain                        |
| text/html                         | html, text/html                         |
| text/xml                          | xml, text/xml                           |
| application/json                  | json, application/json                  |
| application/x-www-form-urlencoded | form, application/x-www-form-urlencoded |
| multipart/form-data               | multipart, multipart/form-data          |
| application/vnd.api+json          | json-api, application/vnd.api+json      |
| application/x-json-stream         | json-stream, application/x-json-stream  |
| application/octet-stream          | octet-stream, application/octet-stream  |
| image/png                         | png, image/png                          |
| image/jpeg                        | jpeg, image/jpeg                        |
| image/jpeg                        | jpg,  image/jpeg                        |
| image/gif                         | gif,  image/gif                         |

## Struct Annotation
```go
type Profile struct {
  // User's email
  Username string `json:"username" required:"true"`
  Status UserStatus `apitype:"int"`
}
type UserStatus int
```

* Comment above the field is being captured as field "description"
* `json:"x"` overrides the field name
* `apitype:"x"` overrides the field type
* `required:"true"` marks the field as required

## Data Types Conversion
Go types are being converted into OpenAPI accepted format

| go Type | Converted Type |
| ------- | -------------- |
| byte    | integer        |
| rune    | integer        |
| int     | integer        |
| int8    | integer        |
| int16   | integer        |
| int32   | integer        |
| int64   | integer        |
| uint    | integer        |
| uint8   | integer        |
| uint16  | integer        |
| uint32  | integer        |
| uint64  | integer        |
| uintptr | integer        |
| float32 | number         |
| float64 | number         |
| bool    | boolean        |

# Tips

## Annotation over Multiple Lines
Last parameter of any tag is could spread over multiple lines, e.g.:
```go
// @desc Lorem ipsum dolor sit amet, consectetur adipiscing
// elit. Nullam rhoncus magna nunc, in faucibus metus pulvinar
// et. Mauris pellentesque enim justo

// @failure 500 {string} Lorem ipsum dolor sit amet, consectetur
// adipiscing elit. Nullam rhoncus magna nunc, in faucibus metus
// pulvinar et. Mauris pellentesque enim justo.
```

## Array References
Use **[]** annotation in from of the reference.
```go
// @body []request.Person

// @success 200 {object} []response.Person OK
```

## And Endpoint With Many Decralarions
An endpoint, with the same URL could be declared separately with different methods, and the generator will properly group them.
### Example
```go
// @summary Get User
// @param id path {int} true User ID
// @produce text
// @success 200 {string} OK
// @router /user/{id} [post]

// ...

// @summary Delete User
// @param id path {int} true User ID
// @produce text
// @success 200 {string} OK
// @router /user/{id} [delete]
```
will produce
```yaml
/user/{id}:
  get:
    summary: Get User
    parameters:
    - name: id
      description: User ID
      in: path
      required: true
      schema:
        type: integer
    responses:
      "200":
        description: OK
        content:
          plain/text:
            schema:
              type: string
  delete:
    summary: Delete User
    parameters:
    - name: id
      description: User ID
      in: path
      required: true
      schema:
        type: integer
    responses:
      "200":
        description: OK
        content:
          plain/text:
            schema:
              type: string
```

# APIDoc CLI
Run the `apidoc -h` to all flags and commands:
```console
API Documentation Generator

Usage:
   [flags]
   [command]

Available Commands:
  help        Help about any command
  version     Show the APIDoc version

Flags:
  -e, --endpoints string   Root endpoints folder (default "./")
  -h, --help               Help for this command
  -m, --main string        Main API documentation file (default "main.go")
  -o, --output string      Documentation output folder (default "docs/api")
  -v, --verbose            Show generation warnings

Use " [command] --help" for more information about a command.
```

# About the Project
This project was inspired by [swaggo/swag](https://github.com/swaggo/swag/), designed mainly to handle our API documentation needs, i.e. add support for response wrappers, generate OpenAPI v3.X documentation. Any feedback, contribution to this project is welcomed.

The project is in a beta phase, therefore there might be major changes in near future, the annotation should stay the same, though.

# Contributing
When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change. 

## Pull Request Process
1. Fork it
2. Create your feature branch (`git checkout -b ft/new-feature-name`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin ft/new-feature-name`)
5. Create new Pull Request

> Please make an issue first if the change is likely to increase.

# License
APIDoc is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/spaceavocado/apidoc/blob/master/LICENSE.txt)

