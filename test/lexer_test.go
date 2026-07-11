package test

import (
	"testing"

	"github.com/CandyCrafts/candy/internal/parser/token"

	"github.com/CandyCrafts/candy/internal/parser/lexer"
	"github.com/k0kubun/pp"
)

func TestLexer(t *testing.T) {
	lex := lexer.New([]byte(`
package::("main");

use::(
    "github.com/CandyCrafts/plugins/db", // import NEW db
)

// global attr
#[lang=custom::("github.com/CandyCrafts/LangEngines/Go@latest")]; // import NEW engine

#[db::sqlite::table::("User")];
go::model User {
    #[db::sqlite::index]; // behavior tag 
    pub Id: strings = go::lib::("github.com/google/uuid")::NewString()
    pub Name: string = "none",
}

#[comoser::file::no_edit::(true)];
go::impl User {
    #[db::go::func::delete_rec];
    pub banned() -> go::type::error,
}
	`), "model.cp")

	for {
		tk := lex.NextToken()
		pp.Println(tk.Kind.String())
		if tk.Kind == token.EOF {
			break
		}
	}
}
