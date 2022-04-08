# Welcome to The Bubble Tea Docs

### When do things run?

The `tea.Model` interface defines `Init`, `Update`, and `View` functions as seen in the Elm architecture.

**Init** — `Init()` is called when the program starts, its role is to fire off initial Commands  

**Update** — `Update()` runs when a command is triggered, this is any function that returns a `tea.Msg`  

**View** — `View()` is called automatically after `Update()` to redraw the program with the updated state.  

Definitely check out the Elm architecture resource above to learn more about how ELM works. Also, hop into our [Slack](https://charm.sh/slack) if you have any questions or want to be part of the community!

## Nested Models

### Data Flow
If you're working with nested models, you don't need to worry about using commands to send data to your main model as the data flows downward. 
This means that the parent knows about the children. 

