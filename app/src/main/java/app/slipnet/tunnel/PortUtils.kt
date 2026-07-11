package app.slipnet.tunnel

import app.slipnet.util.AppLog as Log
import java.net.InetSocketAddress
import java.net.ServerSocket
import java.net.Socket

/**
 * Local TCP port helpers shared by the Go-backed DNS tunnel bridges
 * ([DnsttBridge], [VaydnsBridge]).
 *
 * The Go runtime can keep a listener socket bound for a short while after the
 * client is stopped, so these helpers let a bridge probe / wait for a port
 * before rebinding it.
 */
object PortUtils {

    /**
     * @return true if [port] on 127.0.0.1 cannot currently be bound (i.e. it is
     * still held by another process/listener).
     */
    fun isPortInUse(tag: String, port: Int): Boolean {
        return try {
            ServerSocket().use { serverSocket ->
                serverSocket.reuseAddress = true
                serverSocket.bind(InetSocketAddress("127.0.0.1", port))
                false
            }
        } catch (e: java.net.BindException) {
            true
        } catch (e: Exception) {
            Log.w(tag, "Error checking port $port: ${e.message}")
            true
        }
    }

    /**
     * Poll until [port] becomes bindable or [maxWaitMs] elapses.
     *
     * @return true if the port is free by the time this returns.
     */
    fun waitForPortAvailable(tag: String, port: Int, maxWaitMs: Long = 5000): Boolean {
        val startTime = System.currentTimeMillis()
        while (System.currentTimeMillis() - startTime < maxWaitMs) {
            if (!isPortInUse(tag, port)) {
                return true
            }
            Log.d(tag, "Waiting for port $port to be released...")
            Thread.sleep(50)
        }
        return !isPortInUse(tag, port)
    }

    /**
     * @return true if a TCP connection to [host]:[port] can be established
     * within [timeoutMs], used to confirm the local listener is accepting.
     */
    fun canConnect(tag: String, host: String, port: Int, timeoutMs: Int = 2000): Boolean {
        return try {
            Socket().use { socket ->
                socket.connect(InetSocketAddress(host, port), timeoutMs)
                true
            }
        } catch (e: Exception) {
            Log.w(tag, "Listener verify failed: ${e.message}")
            false
        }
    }
}
