package com.protocolbunker.host.util

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import java.net.Inet4Address
import java.net.NetworkInterface

object NetworkUtils {
    @Volatile
    private var testLanProvider: (() -> String?)? = null

    fun setLanProviderForTests(provider: (() -> String?)?) {
        testLanProvider = provider
    }

    fun findLanIpv4(context: Context? = null): String? {
        testLanProvider?.invoke()?.let { return it }

        findActiveNetworkIpv4(context)?.let { return it }
        return findPreferredInterfaceIpv4()
    }

    private fun findActiveNetworkIpv4(context: Context?): String? =
        runCatching {
            val appContext = context ?: return null
            val connectivity = appContext.getSystemService(ConnectivityManager::class.java) ?: return null
            val network = connectivity.activeNetwork ?: return null
            val capabilities = connectivity.getNetworkCapabilities(network) ?: return null
            if (!capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) &&
                !capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET)
            ) {
                return null
            }

            val linkProperties = connectivity.getLinkProperties(network) ?: return null
            linkProperties.linkAddresses
                .mapNotNull { it.address as? Inet4Address }
                .firstOrNull { !it.isLoopbackAddress && it.isSiteLocalAddress }
                ?.hostAddress
        }.getOrNull()

    private fun findPreferredInterfaceIpv4(): String? {
        val interfaces = NetworkInterface.getNetworkInterfaces()?.toList().orEmpty()
        var bestSiteLocal: Candidate? = null
        var bestFallback: Candidate? = null

        for (networkInterface in interfaces) {
            if (!networkInterface.isUp || networkInterface.isLoopback || networkInterface.isVirtual) {
                continue
            }

            val score = interfacePriority(networkInterface.name)
            if (score <= Int.MIN_VALUE / 4) {
                continue
            }

            val addresses = networkInterface.inetAddresses.toList()
            for (address in addresses) {
                val ipv4 = address as? Inet4Address ?: continue
                if (ipv4.isLoopbackAddress) continue

                val candidate = Candidate(ipv4.hostAddress ?: continue, score)
                if (ipv4.isSiteLocalAddress) {
                    if (bestSiteLocal == null || candidate.score > bestSiteLocal.score) {
                        bestSiteLocal = candidate
                    }
                } else if (bestFallback == null || candidate.score > bestFallback.score) {
                    bestFallback = candidate
                }
            }
        }

        return bestSiteLocal?.address ?: bestFallback?.address
    }

    private fun interfacePriority(name: String?): Int {
        val normalized = name?.lowercase().orEmpty()
        return when {
            normalized.startsWith("wlan") ||
                normalized.startsWith("ap") ||
                normalized.startsWith("swlan") ||
                normalized.contains("softap") -> 100

            normalized.startsWith("rndis") ||
                normalized.startsWith("usb") ||
                normalized.startsWith("eth") -> 90

            normalized.startsWith("tun") ||
                normalized.startsWith("tap") ||
                normalized.startsWith("ppp") ||
                normalized.startsWith("rmnet") ||
                normalized.startsWith("ccmni") ||
                normalized.startsWith("v4-rmnet") ||
                normalized.startsWith("vti") ||
                normalized.startsWith("wg") -> Int.MIN_VALUE

            else -> 10
        }
    }

    private data class Candidate(
        val address: String,
        val score: Int,
    )
}
