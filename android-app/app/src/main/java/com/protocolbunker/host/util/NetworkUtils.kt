package com.protocolbunker.host.util

import java.net.Inet4Address
import java.net.NetworkInterface

object NetworkUtils {
    @Volatile
    private var testLanProvider: (() -> String?)? = null

    fun setLanProviderForTests(provider: (() -> String?)?) {
        testLanProvider = provider
    }

    fun findLanIpv4(): String? {
        testLanProvider?.invoke()?.let { return it }

        val interfaces = NetworkInterface.getNetworkInterfaces() ?: return null
        var fallback: String? = null
        while (interfaces.hasMoreElements()) {
            val networkInterface = interfaces.nextElement()
            if (!networkInterface.isUp || networkInterface.isLoopback || networkInterface.isVirtual) {
                continue
            }

            val addresses = networkInterface.inetAddresses
            while (addresses.hasMoreElements()) {
                val address = addresses.nextElement()
                if (address !is Inet4Address || address.isLoopbackAddress) {
                    continue
                }
                if (address.isSiteLocalAddress) {
                    return address.hostAddress
                }
                if (fallback == null) {
                    fallback = address.hostAddress
                }
            }
        }
        return fallback
    }
}
