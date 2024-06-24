package runtime

import (
	"slices"

	"github.com/goptos/utils"
)

type SignalId int
type EffectId int

// A reactive scope holding each Signal's values and subscribed effects.
type Scope struct {
	values         []any
	effects        []func()
	subscriptions  map[SignalId][]EffectId
	running_effect EffectId
}

// Creates a new Scope with correct initialisation.
func (_self *Scope) New() *Scope {
	var cx = new(Scope)
	cx.running_effect = -1
	cx.subscriptions = make(map[SignalId][]EffectId)
	return cx
}

// Creates a new reactive effect which will automatically subscribe to any signals used within it.
func (_self *Scope) CreateEffect(f func()) {
	_self.effects = append(_self.effects, f)
	var id = (EffectId)(len(_self.effects) - 1)
	utils.Debug("Added effect at address 0x%x with id %d. [Scope.create_effect()]\n",
		&_self.effects[id], id)
	_self.runEffect(id)
}

func (_self *Scope) runEffect(effect_id EffectId) {
	if _self.effects == nil {
		utils.Debug("Scope.effects was nil. [Scope.run_effect()]\n")
		return
	}
	var prev_running_effect = _self.running_effect
	_self.running_effect = effect_id
	utils.Debug("Pushed effect with id %d onto the stack. [Scope.run_effect()]\n", effect_id)
	utils.Debug("  prev_running_effect was %d\n", prev_running_effect)
	utils.Debug("Running effect at address 0x%x. [Scope.run_effect()]\n", &_self.effects[effect_id])
	var effect = _self.effects[effect_id]
	effect()
	_self.running_effect = prev_running_effect
	utils.Debug("Popped effect with id %d off the stack. [Scope.run_effect()]\n", effect_id)
}

func (_self *Scope) createSubscription(signal_id SignalId) {
	if _self.running_effect < 0 {
		utils.Debug("No effects waiting for subscription. [Scope.create_subscription()]\n")
		return
	}
	if _self.subscriptions == nil {
		_self.subscriptions = make(map[SignalId][]EffectId)
		utils.Debug("Initialised subscriptions from nil. [Scope.create_subscription()]\n")
	}
	_, ok := _self.subscriptions[signal_id]
	if ok {
		if slices.Contains(_self.subscriptions[signal_id], _self.running_effect) {
			utils.Debug(
				"Effect with id %d already subscribed to signal with id %d. [Scope.create_subscription()])\n",
				_self.running_effect, signal_id)
			return
		}
	}
	_self.subscriptions[signal_id] = append(_self.subscriptions[signal_id], _self.running_effect)
	utils.Debug(
		"Effect with id %d newly subscribed to signal with id %d. [Scope.create_subscription()]\n",
		_self.running_effect, signal_id)
}

func (_self *Scope) updateSubscribers(signal_id SignalId) {
	if len(_self.effects) == 0 {
		utils.Debug("Scope.effects was empty. [Scope.update_subscribers()]\n")
		return
	}
	if _self.subscriptions == nil {
		utils.Debug("Scope.subscriptions was nil. [Scope.update_subscribers()]\n")
		return
	}
	subscribers, ok := _self.subscriptions[signal_id]
	if !ok {
		utils.Debug("No subscription for signal with id %d. [Scope.update_subscribers()]\n", signal_id)
		return
	}
	if len(subscribers) == 0 {
		utils.Debug("No subscribers to update for Scope.subscriptions[%d]. [Scope.update_subscribers()]\n", signal_id)
		return
	}
	for i := 0; i < len(subscribers); i++ {
		utils.Debug("Updating subscriber with effect id %d as signal with id %d has changed. [Scope.update_subscribers()]\n", i, signal_id)
		_self.runEffect(subscribers[i])
	}
}

// A signal that is aware of it's reactive Scope and pointer to it's value within.
type Signal[T any] struct {
	cx *Scope
	id SignalId
}

// Creates a new Signal with a unique id pointing to it's value within a reactive Scope.
func (_self *Signal[T]) New(cx *Scope, value T) Signal[T] {
	cx.values = append(cx.values, value)
	var id = (SignalId)(len(cx.values) - 1)
	utils.Debug("Added signal at address 0x%x with id %d. [Signal.new()]\n", &cx.values[id], id)
	return Signal[T]{
		cx: cx,
		id: id,
	}
}

// Gets the value for a Signal from it's reactive Scope.
func (_self *Signal[T]) Get() T {
	_self.cx.createSubscription(_self.id)
	return _self.cx.values[_self.id].(T)
}

// Sets the value for a Signal in it's reactive Scope.
func (_self *Signal[T]) Set(value T) {
	_self.cx.values[_self.id] = value
	_self.cx.updateSubscribers(_self.id)
}
