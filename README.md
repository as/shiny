# Shiny

This is an incompatible fork of Shiny that strips away the event-driven constructs, replacing them with channels and CSP. 

# Changes

- AVX2 swizzle for amd64 (also in master branch)
- The event pump is gone. (concurrent)
- All events are sent and recieved via channels (concurrent)
- Bare-bones functionality; no widgets (concurrent)
- Only one os window, called the "screen" (concurrent)