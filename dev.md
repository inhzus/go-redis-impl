# Dev Diary

### client consumer

2019-08-28

`internal/pkg/client/client.go:(Client)consume`

My efforts on timeout:

net.Conn interface supplies the function `SetWriteDeadline` and
`SetReadDeadline`. `SetWriteDeadline` is alright for the case. However,
`SetReadDeadline` does not acts as what I expect. Anyway, if I use the
function to set a read deadline, the reading may interrupt when the data is
on the way from the server to the client, that is, next time when the client
send the new request to the server, it may recognize the last response as the
correspond response. Fatally the case will chain up all subsequent responses
parsed by the consumer. So instead I start a new goroutine to wait for and
parse the response when the client need to read from the connection. When it
is timeout, the consumer will stops waiting for the signal that response is
reached and continues to wait for another request sent to the consumer, but
the goroutine will still hang until the response is reached. In this way, I
manage to prevent from reading the last response.

### Server structure affected by the persistence component

2019-08-29

Before I introduced the persistence component, the project has been over 2k
lines and nearly out of my control. During previous projects, I always complain
about my trapping in some detailed thing. To avoid getting caught again I
stuffed the persistence component directly. Making it destroyed the module
structure definitely from the current view.
I put the `SetMsg channel` into the server entity, so the persistence part can
receive the msg in the channel directly. However, to send the `SetMsg` from
client to channel, client entity needs to hold both data index and data
pointer, which guarantees to send the data idx back to persistence. And the
channel will pass through server, processor and finally put in the client
entity. By the way, putting the channel existing in the entity of client into
the client entity is really ugly. This is the reason that I do not want to
review the project today.

So how to fix the situation?

1. Data entity should hold the index position in the data array. This can
classify the meaning of the data member without any comment.
2. Instead of processing the data alone processor should store some information
flags and let the server load them and process together. 
