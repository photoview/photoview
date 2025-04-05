import { Message } from './SubscriptionsHook'
import { NotificationType } from '../../__generated__/globalTypes'

type MessageStateUpdater = {
    add: (message: Message) => void
    removeKey: (key: string) => void
}

// Singleton pattern for global message handling
class GlobalMessageHandler {
    private messageState: MessageStateUpdater | null = null
    private pendingMessages: Message[] = []
    private isInitialized: boolean = false

    // Connect the handler to React state (called from MessageProvider)
    initialize(messageState: MessageStateUpdater) {
        this.messageState = messageState
        this.isInitialized = true

        // Process any messages that were queued before initialization
        if (this.pendingMessages.length > 0) {
            const messages = [...this.pendingMessages]
            this.pendingMessages = []
            messages.forEach(message => this.add(message))
        }
    }

    // Add a message (works both before and after initialization)
    add(message: Message) {
        if (this.isInitialized && this.messageState) {
            this.messageState.add(message)
        } else {
            this.pendingMessages.push(message)
        }
    }

    // Remove a message by key
    removeKey(key: string) {
        if (this.isInitialized && this.messageState) {
            this.messageState.removeKey(key)
        } else {
            this.pendingMessages = this.pendingMessages.filter(msg => msg.key !== key)
        }
    }

    // Helper to create error messages
    addErrorMessage(header: string, content: string) {
        this.add({
            key: Math.random().toString(26),
            type: NotificationType.Message,
            props: {
                negative: true,
                header,
                content
            }
        })
    }
}

// Export a singleton instance
export const globalMessageHandler = new GlobalMessageHandler()
