package app.slipnet.tunnel

import app.slipnet.util.AppLog as Log
import java.net.Socket
import java.security.SecureRandom
import java.security.cert.X509Certificate
import javax.net.ssl.SNIHostName
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLSocket
import javax.net.ssl.SSLSocketFactory
import javax.net.ssl.TrustManager
import javax.net.ssl.X509TrustManager

/**
 * Shared TLS helpers for tunnel transports.
 *
 * These transports connect to self-signed / StunTLS-style servers where
 * certificate validation is intentionally skipped and the SNI hostname is
 * chosen for domain fronting or DPI bypass rather than for authentication.
 */
object TlsUtils {

    /** Trust managers that accept any certificate chain (no validation). */
    fun trustAllManagers(): Array<TrustManager> = arrayOf(object : X509TrustManager {
        override fun checkClientTrusted(chain: Array<X509Certificate>, authType: String) {}
        override fun checkServerTrusted(chain: Array<X509Certificate>, authType: String) {}
        override fun getAcceptedIssuers(): Array<X509Certificate> = emptyArray()
    })

    /** An [SSLContext] initialised with [trustAllManagers]. */
    fun trustAllSslContext(): SSLContext =
        SSLContext.getInstance("TLS").apply { init(null, trustAllManagers(), SecureRandom()) }

    /** An [SSLSocketFactory] that trusts all certificates. */
    fun trustAllSslFactory(): SSLSocketFactory = trustAllSslContext().socketFactory

    /**
     * Wrap [rawSocket] in a TLS layer with the given [sniHost], perform the
     * handshake and return the connected [SSLSocket]. Certificates are not
     * validated (see [trustAllManagers]).
     *
     * @param tag Log tag of the caller.
     */
    fun upgradeToTls(rawSocket: Socket, sniHost: String, port: Int, tag: String): SSLSocket {
        val sslSocket = trustAllSslFactory().createSocket(rawSocket, sniHost, port, true) as SSLSocket
        val params = sslSocket.sslParameters
        params.serverNames = listOf(SNIHostName(sniHost))
        sslSocket.sslParameters = params
        sslSocket.startHandshake()
        Log.i(tag, "TLS handshake complete (${sslSocket.session.protocol}, ${sslSocket.session.cipherSuite})")
        return sslSocket
    }
}
