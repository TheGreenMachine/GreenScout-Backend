# Themes

Inside of the [themes folder](../run/themes/), you may find the CSS files for
each theme. Each theme lives inside a single file CSS. The name of the theme
will be inherited from the name of the file (omitting '.css').

The themes will be loaded via a link HTML tag, so expect the entire CSS file to
have an effect on the site.

ALL CSS variables inside a theme must be followed by `!important`. This will let
the browser know to use our CSS variable definitions instead the defaults from
GreenScoutJS.

```css
--font-color: #1e2226 !important;
```

If a theme does not provide a definition for a CSS variable, its default will be
used in place. The default theme can be found
[here in the GreenScoutJS repo. **Reference this file.**](https://github.com/TheGreenMachine/GreenScoutJS/blob/main/greenscoutjs/src/index.css)
**for all available CSS variables to override.** Again, make sure to include
`!important` when copy and pasting.
