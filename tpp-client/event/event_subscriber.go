package event

type EventSubscriber interface {

	// SubscribeToMintRequestEvent subscribes to the MintRequest event.
	// Returns `true` if a new subscription was created, `false` if it already existed.
	SubscribeToMintRequestEvent() (bool, error)

	// UnsubscribeFromMintRequestEvent unsubscribes from the MintRequest event.
	// Returns `true` if there was a subscription and terminated, `false` otherwise.
	UnsubscribeFromMintRequestEvent() bool

	// SubscribeToAuthGrantedEvent subscribes to the AuthGranted event.
	// Returns `true` if a new subscription was created, `false` if it already existed.
	SubscribeToAuthGrantedEvent() (bool, error)

	// UnsubscribeFromAuthGrantedEvent unsubscribes from the AuthGranted event.
	// Returns `true` if there was a subscription and terminated, `false` otherwise.
	UnsubscribeFromAuthGrantedEvent() bool
}
