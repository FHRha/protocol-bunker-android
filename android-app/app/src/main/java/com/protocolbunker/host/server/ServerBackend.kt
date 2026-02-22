package com.protocolbunker.host.server

import android.content.Context

internal interface ServerBackend {
    val name: String

    suspend fun start(
        context: Context,
        port: Int,
        devMode: Boolean,
        logger: (String) -> Unit,
        onProcessTerminated: ((Int) -> Unit)? = null
    )

    suspend fun stop(logger: (String) -> Unit)
}
