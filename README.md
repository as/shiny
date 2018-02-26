# Shiny

This is an incompatible fork of Shiny that strips away the event-driven constructs, replacing them with channels and CSP. 

# Changes

- The event pump is gone. 
- All events are sent and recieved via channels
- Bare-bones functionality; no widgets
- AVX2 swizzle for amd64
- Only one os window, called the "screen"