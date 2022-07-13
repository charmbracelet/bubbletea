# Creating Nested Models

There may be situations where you want to have your own nested models for your project. 
This can allow you to toggle between different views and organize your logic for `Update`.

## Showing Specific Nested Models
In Bubble Tea programs, you can decide which components are shown by holding a `state` field in your main model struct. 
[Check out our nested model example](https://github.com/charmbracelet/bubbletea/tree/master/examples/nested-models/)
As you can see, the main impact the `state` has is in the `View` method.
