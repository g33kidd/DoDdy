package com.github.dod.doddy.core

import net.dv8tion.jda.core.events.Event
import net.dv8tion.jda.core.events.message.MessageReceivedEvent
import net.dv8tion.jda.core.hooks.EventListener

class MessageListener(private val modules: Modules) : EventListener {
    override fun onEvent(event: Event) {
        print(event)
        when (event) {
            is MessageReceivedEvent -> onMessageReceived(event)
        }
    }

    private fun onMessageReceived(event: MessageReceivedEvent) {
        if (event.author.isBot) return
        val content = event.message.contentRaw
        val length = content.length
        if (length == 0) return
        if (content[0] == '!') {//TODO: use from db
            val parsedMessage = content.slice(1 until length - 1).trim().split("/ +/g")
            val parsedMessageSize = parsedMessage.size
            if (parsedMessageSize == 0) return
            val commandName = parsedMessage[0]
            val commandArgs = if (parsedMessageSize > 1) { // has args
                parsedMessage.subList(1, parsedMessageSize - 1)
            } else {
                emptyList()
            }
            modules.onCommand(commandName, event, commandArgs)
        }
    }
}