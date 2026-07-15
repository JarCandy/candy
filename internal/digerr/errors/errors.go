package errors

import (
	"github.com/CandyCrafts/candy/internal/digerr/translate"
	"github.com/CandyCrafts/candy/internal/parser/token"
)

type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

const (
	testErrorIndex = iota
	lexerIllegalIndex
	lexerNoClosingIndex
	errorLoadLibraryIndex
	executingCommandsIndex
	parsingErrorIndex
	errorActionMapIndex
)

func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "error"
	}
}

// Не менять массив!!
var Errors = []Error{
	define(0, "TestError", SeverityWarning,
		text("test diagnostic error", "тестовая ошибка для проверки механизма диагностики"),
		nil,
		text("test error must not get into release builds", "тест-ошибка не должна попадать в релиз!!"),
	),
	define(1, "LexerIllegal", SeverityError,
		text("illegal character", "недопустимый символ"),
		nil,
		text("the character was not recognized; it may be unsupported or a typo", "символ не распознан, возможно он не поддерживается или это опечатка"),
	),
	define(2, "LexerNoClosing", SeverityError,
		text("missing closing character", "не найден закрывающий символ"),
		text("Close it!", "Закрой за собой!"),
		text("the opening character was not closed before the end of file; it may be missing or mistyped", "открывающий символ не был закрыт до конца файла, возможно пропущен или это опечатка"),
		text("check that every opening character, such as quotes or brackets, has a matching closing character", "проверьте, что все открывающие символы (например, кавычки, скобки) имеют соответствующие закрывающие символы"),
		text("if the error appears inside a string, check escaping inside the string", "если ошибка возникает внутри строки, проверьте правильность экранирования символов внутри строки"),
		text("you may have forgotten to close a string or comment; check the matching delimiters", "возможно вы забыли закрыть строку или комментарий, проверьте соответствующие символы в коде"),
	),
	define(3, "ErrorLoadLibrary", SeverityError,
		text("failed to load library", "не удалось загрузить библиотеку"),
		nil,
		text("failed to load library. Cause:", "не удалось загрузить библиотеку. Причина ошибки:"),
	),
	define(4, "ExecutingCommands", SeverityError,
		text("failed to execute command", "не удалось выполнить команду"),
		nil,
		text("command execution error:", "ошибка выполнения команды: "),
	),
	define(5, "ParsingError", SeverityError,
		text("parse error", "ошибка разбора"),
		nil,
		text("expected: got:", "ожидалось: было получено:"),
	),
	define(6, "ErrorActionMap", SeverityWarning,
		text("action map error", "ошибка работы с таблицей"),
		nil,
		text("create https://github.com/fugalang/fugu/issues", "создайте https://github.com/fugalang/fugu/issues"),
	),
}

type Error struct {
	Code          uint16
	Severity      Severity
	CodeName      string // название ошибки, для удобства
	Message       string // сообщение об ошибке, кратокое описание
	Arrow         string // строка с указанием места ошибки и пояснением.
	IsShowSnippet bool
	Description   []string

	MessageTranslations     translate.Translations
	ArrowTranslations       translate.Translations
	DescriptionTranslations []translate.Translations

	Start uint64
	End   uint64
	Pos   token.Position
}

func (e Error) Update(tk token.Token) Error {
	e.Start = tk.Start
	e.End = tk.End
	e.Pos = tk.Pos
	return e.Localize()
}

func (e Error) IU(fileModule string, description []string) Error {
	e = e.Localize()
	e.Description = description
	e.Pos.FileName = fileModule
	e.Pos.Line = 0
	return e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Localize().Message
}

func (e Error) Localize() Error {
	if len(e.MessageTranslations) > 0 {
		e.Message = translate.Resolve(e.MessageTranslations)
	}
	if len(e.ArrowTranslations) > 0 {
		e.Arrow = translate.Resolve(e.ArrowTranslations)
	}
	if len(e.DescriptionTranslations) > 0 {
		e.Description = make([]string, 0, len(e.DescriptionTranslations))
		for _, desc := range e.DescriptionTranslations {
			e.Description = append(e.Description, translate.Resolve(desc))
		}
	}
	return e
}

func define(code uint16, codeName string, severity Severity, message translate.Translations, arrow translate.Translations, description ...translate.Translations) Error {
	return Error{
		Code:                    code,
		CodeName:                codeName,
		Severity:                severity,
		MessageTranslations:     message,
		ArrowTranslations:       arrow,
		DescriptionTranslations: description,
		IsShowSnippet:           len(arrow) > 0,
	}.Localize()
}

func text(eng string, ru string) translate.Translations {
	return translate.Translations{
		{Language: translate.DefaultLanguage, Text: eng},
		{Language: "ru", Text: ru},
	}
}

func LexerUnexpectedLess(tk token.Token) Error {
	return lexerIllegal(
		tk,
		text("Unexpected '<'", "Неожиданный символ '<'"),
		text("use '<-' for a reverse transition", "используйте '<-' для обратного перехода"),
	)
}

func LexerUnexpectedSharp(tk token.Token) Error {
	return lexerIllegal(
		tk,
		text("Unexpected '#'", "Неожиданный символ '#'"),
		text("an attribute must start with '#['", "атрибут должен начинаться с '#['"),
	)
}

func LexerNestedAttribute(tk token.Token) Error {
	return lexerIllegal(
		tk,
		text("Nested attributes are not supported", "Вложенные атрибуты не поддерживаются"),
		text("close the current attribute before starting another bracket block", "закройте текущий атрибут перед началом нового блока скобок"),
	)
}

func LexerUnknownCharacter(tk token.Token) Error {
	return lexerIllegal(
		tk,
		text("Unknown character", "Неизвестный символ"),
		text("the character was not recognized by the lexer", "символ не распознан лексером"),
	)
}

func LexerUnterminatedMultilineComment(tk token.Token) Error {
	return lexerNoClosing(
		tk,
		text("Unterminated multiline comment", "Не закрыт многострочный комментарий"),
		text("add the closing */ delimiter", "добавьте закрывающий символ */"),
	)
}

func LexerUnterminatedString(tk token.Token) Error {
	return lexerNoClosing(
		tk,
		text("Unterminated string", "Не закрыта строка"),
		text("add the closing double quote", "добавьте закрывающую кавычку \""),
	)
}

func LexerUnterminatedRawString(tk token.Token) Error {
	return lexerNoClosing(
		tk,
		text("Unterminated raw string", "Не закрыта raw-строка"),
		text("add the closing ` delimiter", "добавьте закрывающий символ `"),
	)
}

func LexerInvalidCharacterLiteral(tk token.Token) Error {
	return lexerIllegal(
		tk,
		text("Invalid character literal", "Некорректный символьный литерал"),
		text("a character literal must contain one character and end with an apostrophe", "символьный литерал должен содержать один символ и закрываться апострофом"),
	)
}

func ParserMissingClosingParen(tk token.Token) Error {
	return parsingError(
		tk,
		text("Expected closing parenthesis", "Ожидалась закрывающая скобка"),
		text("a parenthesized expression must end with ')'", "выражение в скобках должно завершаться символом ')'"),
	)
}

func ParserUnexpectedExprToken(tk token.Token) Error {
	return parsingError(
		tk,
		text("Unexpected expression token", "Неожиданный токен в выражении"),
		text("expected a literal, identifier, unary minus, or parenthesized expression", "ожидался литерал, идентификатор, унарный минус или выражение в скобках"),
	)
}

func ParserAttrPathSegment(tk token.Token) Error {
	return parsingError(
		tk,
		text("Expected attribute path segment", "Ожидался сегмент пути атрибута"),
		text("expected an identifier in the attribute access path", "ожидался идентификатор в пути доступа к атрибуту"),
		text("an attribute path must look like db::sqlite or db::sqlite(...)", "путь атрибута должен выглядеть как db::sqlite или db::sqlite(...)"),
	)
}

func ParserArg(tk token.Token) Error {
	return parsingError(
		tk,
		text("Argument could not be parsed", "Аргумент не разобран"),
		text("failed to parse an attribute call argument", "не удалось разобрать аргумент вызова атрибута"),
		text("an argument must be a literal, named literal name: value, or module::item(...) access", "аргумент должен быть литералом, именованным литералом name: value или доступом module::item(...)"),
	)
}

func ParserArgSeparator(tk token.Token) Error {
	return parsingError(
		tk,
		text("Expected comma or closing parenthesis", "Ожидалась запятая или закрывающая скобка"),
		text("after an argument, expected a comma for the next argument or a closing parenthesis", "после аргумента ожидалась запятая для следующего аргумента или закрывающая скобка"),
	)
}

func ParserAttrAccess(tk token.Token) Error {
	return parsingError(
		tk,
		text("Attribute access could not be parsed", "Доступ к атрибуту не разобран"),
		text("failed to parse the argument value as attribute access", "не удалось разобрать значение аргумента как доступ к атрибуту"),
	)
}

func ParserArgValue(tk token.Token) Error {
	return parsingError(
		tk,
		text("Expected argument value", "Ожидалось значение аргумента"),
		text("expected a literal or attribute access like module::item(...)", "ожидался литерал или доступ к атрибуту вида module::item(...)"),
		text("strings, numbers, characters, bool values, and identifiers are supported", "поддерживаются строки, числа, символы, bool-значения и идентификаторы"),
	)
}

func ParserUnexpectedTopLevel(tk token.Token) Error {
	return parsingError(
		tk,
		text("Unexpected top-level token", "Неожиданный токен верхнего уровня"),
		text("expected package, use, a declaration, pub modifier, let, or an attribute at the top level", "на верхнем уровне ожидались package, use, объявление, модификатор pub, let или атрибут"),
	)
}

func lexerIllegal(tk token.Token, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[lexerIllegalIndex], tk, arrow, description...)
}

func lexerNoClosing(tk token.Token, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[lexerNoClosingIndex], tk, arrow, description...)
}

func parsingError(tk token.Token, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[parsingErrorIndex], tk, arrow, description...)
}

func diagnostic(err Error, tk token.Token, arrow translate.Translations, description ...translate.Translations) Error {
	err = err.Update(tk)
	err.IsShowSnippet = true
	err.ArrowTranslations = arrow
	if len(description) > 0 {
		err.DescriptionTranslations = description
	}
	return err.Localize()
}
