package errors

import (
	digreyt "github.com/rp1s/digreyt"
	"github.com/rp1s/digreyt/translate"
)

type Error = digreyt.Error
type Span = digreyt.Span
type Position = digreyt.Position
type Severity = digreyt.Severity

const (
	SeverityError   = digreyt.SeverityError
	SeverityWarning = digreyt.SeverityWarning
	SeverityInfo    = digreyt.SeverityInfo
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

func text(eng string, ru ...string) translate.Translations {
	values := translate.Translations{
		{Language: translate.DefaultLanguage, Text: eng},
	}
	if len(ru) > 0 && ru[0] != "" {
		values = append(values, translate.Translation{Language: "ru", Text: ru[0]})
	}
	return values
}

func LexerUnexpectedLess(span Span) Error {
	return lexerIllegal(
		span,
		text("Unexpected '<'", "Неожиданный символ '<'"),
		text("use '<-' for a reverse transition", "используйте '<-' для обратного перехода"),
	)
}

func LexerUnexpectedSharp(span Span) Error {
	return lexerIllegal(
		span,
		text("Unexpected '#'", "Неожиданный символ '#'"),
		text("an attribute must start with '#['", "атрибут должен начинаться с '#['"),
	)
}

func LexerNestedAttribute(span Span) Error {
	return lexerIllegal(
		span,
		text("Nested attributes are not supported", "Вложенные атрибуты не поддерживаются"),
		text("close the current attribute before starting another bracket block", "закройте текущий атрибут перед началом нового блока скобок"),
	)
}

func LexerUnknownCharacter(span Span) Error {
	return lexerIllegal(
		span,
		text("Unknown character", "Неизвестный символ"),
		text("the character was not recognized by the lexer", "символ не распознан лексером"),
	)
}

func LexerUnterminatedMultilineComment(span Span) Error {
	return lexerNoClosing(
		span,
		text("Unterminated multiline comment", "Не закрыт многострочный комментарий"),
		text("add the closing */ delimiter", "добавьте закрывающий символ */"),
	)
}

func LexerUnterminatedString(span Span) Error {
	return lexerNoClosing(
		span,
		text("Unterminated string", "Не закрыта строка"),
		text("add the closing double quote", "добавьте закрывающую кавычку \""),
	)
}

func LexerUnterminatedRawString(span Span) Error {
	return lexerNoClosing(
		span,
		text("Unterminated raw string", "Не закрыта raw-строка"),
		text("add the closing ` delimiter", "добавьте закрывающий символ `"),
	)
}

func LexerInvalidCharacterLiteral(span Span) Error {
	return lexerIllegal(
		span,
		text("Invalid character literal", "Некорректный символьный литерал"),
		text("a character literal must contain one character and end with an apostrophe", "символьный литерал должен содержать один символ и закрываться апострофом"),
	)
}

func ParserMissingClosingParen(span Span) Error {
	return parsingError(
		span,
		text("Expected closing parenthesis", "Ожидалась закрывающая скобка"),
		text("a parenthesized expression must end with ')'", "выражение в скобках должно завершаться символом ')'"),
	)
}

func ParserUnexpectedExprToken(span Span) Error {
	return parsingError(
		span,
		text("Unexpected expression token", "Неожиданный токен в выражении"),
		text("expected a literal, identifier, unary minus, or parenthesized expression", "ожидался литерал, идентификатор, унарный минус или выражение в скобках"),
	)
}

func ParserAttrPathSegment(span Span) Error {
	return parsingError(
		span,
		text("Expected attribute path segment", "Ожидался сегмент пути атрибута"),
		text("expected an identifier in the attribute access path", "ожидался идентификатор в пути доступа к атрибуту"),
		text("an attribute path must look like db::sqlite or db::sqlite(...)", "путь атрибута должен выглядеть как db::sqlite или db::sqlite(...)"),
	)
}

func ParserArg(span Span) Error {
	return parsingError(
		span,
		text("Argument could not be parsed", "Аргумент не разобран"),
		text("failed to parse an attribute call argument", "не удалось разобрать аргумент вызова атрибута"),
		text("an argument must be a literal, named literal name: value, or module::item(...) access", "аргумент должен быть литералом, именованным литералом name: value или доступом module::item(...)"),
	)
}

func ParserArgSeparator(span Span) Error {
	return parsingError(
		span,
		text("Expected comma or closing parenthesis", "Ожидалась запятая или закрывающая скобка"),
		text("after an argument, expected a comma for the next argument or a closing parenthesis", "после аргумента ожидалась запятая для следующего аргумента или закрывающая скобка"),
	)
}

func ParserAttrAccess(span Span) Error {
	return parsingError(
		span,
		text("Attribute access could not be parsed", "Доступ к атрибуту не разобран"),
		text("failed to parse the argument value as attribute access", "не удалось разобрать значение аргумента как доступ к атрибуту"),
	)
}

func ParserArgValue(span Span) Error {
	return parsingError(
		span,
		text("Expected argument value", "Ожидалось значение аргумента"),
		text("expected a literal or attribute access like module::item(...)", "ожидался литерал или доступ к атрибуту вида module::item(...)"),
		text("strings, numbers, characters, bool values, and identifiers are supported", "поддерживаются строки, числа, символы, bool-значения и идентификаторы"),
	)
}

func ParserUnexpectedTopLevel(span Span) Error {
	return parsingError(
		span,
		text("Unexpected top-level token", "Неожиданный токен верхнего уровня"),
		text("expected package, use, a declaration, pub modifier, let, or an attribute at the top level", "на верхнем уровне ожидались package, use, объявление, модификатор pub, let или атрибут"),
	)
}

func ParserPackagePath(span Span) Error {
	return parsingError(
		span,
		text("Package path is too long", "Слишком длинный путь package"),
		text("a package declaration can contain only one path segment", "объявление package может содержать только один сегмент пути"),
		text("use package::(...) instead of package::module::(...)", "используйте package::(...) вместо package::module::(...)"),
	)
}

func ParserAttrStart(span Span) Error {
	return parsingError(
		span,
		text("Expected attribute path", "Ожидался путь атрибута"),
		text("an attribute or access expression must start with an identifier", "атрибут или выражение доступа должно начинаться с идентификатора"),
	)
}

func ParserAttrsStart(span Span) Error {
	return parsingError(
		span,
		text("Expected attribute block", "Ожидался блок атрибута"),
		text("an attribute block must start with '#['", "блок атрибута должен начинаться с '#['"),
	)
}

func ParserAttrsSeparator(span Span) Error {
	return parsingError(
		span,
		text("Expected comma or closing attribute bracket", "Ожидалась запятая или закрывающая скобка атрибута"),
		text("after an attribute entry, expected ',' or ']'", "после элемента атрибута ожидалась ',' или ']'"),
	)
}

func ParserAttrsClosing(span Span) Error {
	return parsingError(
		span,
		text("Expected closing attribute bracket", "Ожидалась закрывающая скобка атрибута"),
		text("an attribute block must end with ']'", "блок атрибута должен завершаться ']'"),
	)
}

func ParserTypeStart(span Span) Error {
	return parsingError(
		span,
		text("Expected type path", "Ожидался путь типа"),
		text("a type must start with an identifier", "тип должен начинаться с идентификатора"),
	)
}

func ParserTypePathSegment(span Span) Error {
	return parsingError(
		span,
		text("Expected type path segment", "Ожидался сегмент пути типа"),
		text("expected an identifier after '::' in the type path", "после '::' в пути типа ожидался идентификатор"),
		text("a type path must look like Type or module::Type", "путь типа должен выглядеть как Type или module::Type"),
	)
}

func ParserLetStart(span Span) Error {
	return parsingError(
		span,
		text("Expected let declaration", "Ожидалось объявление let"),
		text("a variable declaration must start with 'let' or 'pub let'", "объявление переменной должно начинаться с 'let' или 'pub let'"),
	)
}

func ParserLetName(span Span) Error {
	return parsingError(
		span,
		text("Expected variable name", "Ожидалось имя переменной"),
		text("after 'let', expected an identifier variable name", "после 'let' ожидался идентификатор имени переменной"),
	)
}

func ParserLetBody(span Span) Error {
	return parsingError(
		span,
		text("Expected variable type or value", "Ожидался тип или значение переменной"),
		text("after the variable name, expected ': Type', '= value', or both", "после имени переменной ожидалось ': Type', '= value' или оба элемента"),
	)
}

func ParserLetValue(span Span) Error {
	return parsingError(
		span,
		text("Expected variable value", "Ожидалось значение переменной"),
		text("after '=', expected an expression for the variable value", "после '=' ожидалось выражение значения переменной"),
	)
}

func ParserArgsStart(span Span) Error {
	return parsingError(
		span,
		text("Expected argument list", "Ожидался список аргументов"),
		text("an argument list must start with '('", "список аргументов должен начинаться с '('"),
	)
}

func ParserArgsClosingParen(span Span) Error {
	return parsingError(
		span,
		text("Expected closing parenthesis for arguments", "Ожидалась закрывающая скобка аргументов"),
		text("an argument list must end with ')'", "список аргументов должен завершаться ')'"),
	)
}

func ParserOptionalSemicolon(span Span) Error {
	return withSeverity(diagnostic(
		Errors[parsingErrorIndex],
		span,
		text("Optional semicolon is missing", "Необязательная точка с запятой отсутствует"),
		text("put ';' after the declaration to make the boundary explicit", "поставьте ';' после объявления, чтобы явно отделить его от следующей конструкции"),
	), SeverityWarning)
}

func ParserUseStart(span Span) Error {
	return parsingError(
		span,
		text("Expected use import list", "Ожидался список импортов use"),
		text("a use declaration must continue with '('", "объявление use должно продолжаться символом '('"),
	)
}

func ParserUsePath(span Span) Error {
	return parsingError(
		span,
		text("Expected import path", "Ожидалась ссылка импорта"),
		text("a use import entry must start with a string path", "элемент use должен начинаться со строковой ссылки"),
	)
}

func ParserUseAlias(span Span) Error {
	return parsingError(
		span,
		text("Expected import alias", "Ожидался алиас импорта"),
		text("after '->', expected an identifier alias", "после '->' ожидался идентификатор алиаса"),
	)
}

func ParserUseSeparator(span Span) Error {
	return parsingError(
		span,
		text("Expected comma or closing parenthesis", "Ожидалась запятая или закрывающая скобка"),
		text("after a use import entry, expected ',' or ')'", "после элемента use ожидалась ',' или ')'"),
	)
}

func ParserUseClosingParen(span Span) Error {
	return parsingError(
		span,
		text("Expected closing parenthesis for use", "Ожидалась закрывающая скобка use"),
		text("a use import list must end with ')'", "список импортов use должен завершаться ')'"),
	)
}

func lexerIllegal(span Span, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[lexerIllegalIndex], span, arrow, description...)
}

func lexerNoClosing(span Span, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[lexerNoClosingIndex], span, arrow, description...)
}

func parsingError(span Span, arrow translate.Translations, description ...translate.Translations) Error {
	return diagnostic(Errors[parsingErrorIndex], span, arrow, description...)
}

func diagnostic(err Error, span Span, arrow translate.Translations, description ...translate.Translations) Error {
	err = err.Update(span)
	err.IsShowSnippet = true
	err.ArrowTranslations = arrow
	if len(description) > 0 {
		err.DescriptionTranslations = description
	}
	return err.Localize()
}

func withSeverity(e Error, severity Severity) Error {
	e.Severity = severity
	return e
}
