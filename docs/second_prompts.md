Add these improvements:
* Add auto migration of MongoDB tables. For example: “posts” table must be created if not created during the application start.
* Use gin framework instead of Mux
* Move handlers to “http” directory
* Move routes from main.go to separate “http/route.go” file
* Move “Load templates with custom functions” to “web” directory
* Rename “handler” package to “controller” and rename PostHandler to PostController
* Create “internal/db” folder with mongoldb/go file and move database initialization there
* All public methods/function must have go-doc comments
* Use standard “slog” package for logging where needed
* Add docker-compose.yaml file with Mongodb and Galang-server containers and use docker-compose in the “docker-up” make command
* Fix bug: “Create post” form validation happens on each field change instead of only on clicking the “Create Post” button. Please set create/update forms validation only on the submit button click.
