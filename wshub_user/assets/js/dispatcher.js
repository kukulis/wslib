export class Dispatcher {
    constructor() {
        this.listeners = new Map();
    }

    addListener(eventName, listener) {
        if (!this.listeners.has(eventName)) {
            this.listeners.set(eventName, []);
        }

        let eventListeners = this.listeners.get(eventName);
        eventListeners.push(listener);

        this.listeners.set(eventName, eventListeners);
    }

    dispatch(eventName, parameters) {
        if (!this.listeners.has(eventName)) {
            return null;
        }

        let result = null;
        for (let listener of this.listeners.get(eventName)) {
            result = listener(parameters);
        }

        return result;
    }
}

// export const mainDispatcher = new Dispatcher()