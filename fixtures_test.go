package infuse_test

var threeHandlerFixture = `start first
attempting next for first
start second
attempting next for second
start third
attempting next for third
no next for third
end third
finished next for second
end second
finished next for first
end first
`

var complexHandlerFixture = `start first-first
attempting next for first-first
start first-second
attempting next for first-second
start second-first
attempting next for second-first
start second-second
attempting next for second-second
no next for second-second
end second-second
finished next for second-first
end second-first
finished next for first-second
end first-second
finished next for first-first
end first-first
`

var branchedHandlerFixture = `start base
attempting next for base
start first
attempting next for first
start base
attempting next for base
start second
attempting next for second
no next for second
end second
finished next for base
end base
finished next for first
end first
finished next for base
end base
`

var multipleCallHandlerFixture = `start first
attempting next for first
start second
attempting next for second
no next for second
attempting next for second
no next for second
attempting next for second
no next for second
end second
finished next for first
attempting next for first
start second
attempting next for second
no next for second
attempting next for second
no next for second
attempting next for second
no next for second
end second
finished next for first
end first
`

var panicRecoveryHandlerFixture = `start first
attempting next for first
start recoverable
start second
attempting next for second
start third
attempting next for third
panicking
recovered and attempting next
start second
attempting next for second
start third
attempting next for third
already panicked
finished next for third
end third
finished next for second
end second
finished next after recovery
finished next for first
end first
`

var stackedHandlerFixture = `start first stacked
end first stacked
start second stacked
end second stacked
start third stacked
end third stacked
start fourth handled
attempting next for fourth handled
start fifth stacked
end fifth stacked
finished next for fourth handled
end fourth handled
`
