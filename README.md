GORM Repository
===============

Gormrepo is a repository generator for Gorm.

# Install

``` bash
$ go get github.com/l-vitaly/cmd/gormrepogen/...
```

For example, given this struct: 

``` golang
package model

type User struct {
    gorm.Model
    FirstName string
    LastName string
    Birthday time.Time
}
```

running this command: 

``` bash
$ gormrepogen -t=User
```

in the same directory will create the file user_base_repo.go, in package model and struct userBaseRepo.

Typically this process would be run using go generate, like this:

//go:generate gormrepogen -t=User

# Example 

``` golang
package main

import (
	"github.com/jinzhu/gorm"
	"github.com/l-vitaly/gormrepo"
)

type UserRepo struct {
    userBaseRepo
}

func NewUserRepo(db *gorm.DB) *UserRepo {
    return &UserRepo{userBaseRepo{Repo: db}}
}

func main() {
    db, err := gorm.Open("postgres", "postgres://localhost:5432/dbname")
	if err != nil {
		log.Println("err", err)
		os.Exit(1)
	}
	defer db.Close()
	
	userRepo := NewUserRepo(db)
	
	// use repo here
}
```

# Available Criteria

And(query interface{}, args ...interface{}) CriteriaOption

Not(query interface{}, args ...interface{}) CriteriaOption

Or(query interface{}, args ...interface{}) CriteriaOption

Select(columns interface{}, args ...interface{}) CriteriaOption

OrderBy(name string, orientation string, reorder bool) CriteriaOption

Limit(limit int) CriteriaOption

Offset(offset int) CriteriaOption

Preload(field string) CriteriaOption

# Available Methods

Related(claim *T, related interface{}, criteria ...gormrepo.CriteriaOption) (*T, error)

Get(id uint) (*T, error)

GetAll() ([]*T, error)

GetBy(criteria ...gormrepo.CriteriaOption) ([]*T, error)

GetByFirst(criteria ...gormrepo.CriteriaOption) (*T, error)

GetByLast(criteria ...gormrepo.CriteriaOption) (*T, error)

Create(entity T) (*T, error)

Update(entity *T, fields gormrepo.Fields, criteria ...gormrepo.CriteriaOption) (*T, error)

AutoMigrate() error

AddUniqueIndex(name string, columns ...string)

AddForeignKey(field string, dest string, onDelete string, onUpdate string) error

AddIndex(name string, columns ...string) error
