# Service (Application) Shell

TL;DR - A reference library demonstrating a pattern of decoupling a _service_ from the _server_ entities.
Read the test for implementation detials. Continue reading for context.

## Motivation

This will only ever be reference code. To these ends, use this code to start a conversation about
service operation and development. They are different concerns that get expressed as code.

If this is you're situation you probably need to tease a part the server from the service.

- Operating your code in containers and gracefully shutting down SIGHUP, SIGINT, SIGTERM
- You've been tasked to write a service! This is the eleventh time you've implemented a logger and telemetry.
- Pager went off! A service is experiencing lower throughput. Did you remember to set MaxHeaderBytes?
- You've been paged (again) because the service is experience has degraded throughput. Is Read(Write)Timeout set?

If you've experienced anything like this, I believe you have found the edges of the bounded contexts between your
organization's operational and development responsibilities. This kills velocity. It results in copypasta bugs. It is
difficult to keep up with the best current way to do things.

It is also a sign that your operational domain has traction. The more distributed services the greater the need for
thoughtful touchpoints between _how_ code is served and _what_ is served. People are by far, are the most important
element. This reference code is not useful in the least without meaningful conversation between the actors responsible
for the how and the what to see that servers have become distinct from services.

### Why write this

This code exists so operations and development teams are motivated to write the code most appropriate for their context.
Telemetry and logging are great candidates to be written and maintained in one place and consumed as a black box.
Knowing how to drain connections gracefully as well. This attempt to operationalize this pattern is just that - one way
of doing it. May you only ever have to implement a logger or tracing solution once.

### What even is restraint

- Request-scoped logging. A [reasonable use case](https://dave.cheney.net/tag/logging) which is specific to your
use case. Implement it as you see fit.

- Request-scoped tracing. Similar sentiment, however consider leveraging the [ServeHTTP implementation](shell.go#L117)
for injecting a tracer, use a global tracing facility or try a different model not listed.

#### Why use the actor model

If you are reading the source you will see there is a single dependency - [run](https://github.com/oklog/run). You won't
find it in the standard library but I reach for this library pretty often because I think the Actor model works well
with servers I write, YMMV.
