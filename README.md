# Candy


```
package("main");

use (
    "github.com/CandyCrafts/plugins/std",
)

#[lang=custom("github.com/CandyCrafts/LangEngines/Go@latest")]

go::struct Name {
    pub (
        name: type = expr,
    )
}
```

> Many decisions in the code were made with the understanding that sacrificing readability and logical structure for performance gains would not be worthwhile, as those gains would have little impact on the tools user experience.
