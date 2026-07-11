package app.slipnet.widget

import android.appwidget.AppWidgetManager
import android.content.ComponentName
import android.content.Context
import android.widget.RemoteViews
import app.slipnet.domain.model.ConnectionState
import app.slipnet.service.VpnConnectionManager

/**
 * Shared plumbing for the app-widget providers ([VpnWidgetProvider],
 * [VpnWidgetCompactProvider]).
 *
 * Each provider only differs in the layout it renders; the boilerplate for
 * resolving the current connection state and pushing [RemoteViews] to every
 * live widget instance is identical, so it lives here.
 */
object WidgetUpdater {

    /**
     * The current VPN state, falling back to [ConnectionState.Disconnected]
     * when the (lateinit) manager has not been injected yet.
     */
    fun stateOrDisconnected(connectionManager: VpnConnectionManager?): ConnectionState =
        connectionManager?.connectionState?.value ?: ConnectionState.Disconnected

    /** Render and push [state] to each id in [appWidgetIds]. */
    fun updateWidgets(
        appWidgetManager: AppWidgetManager,
        appWidgetIds: IntArray,
        state: ConnectionState,
        buildViews: (ConnectionState) -> RemoteViews
    ) {
        for (appWidgetId in appWidgetIds) {
            appWidgetManager.updateAppWidget(appWidgetId, buildViews(state))
        }
    }

    /**
     * Push [state] to every live instance of the widget backed by [providerClass].
     * No-op when the widget manager is unavailable or no instances exist.
     */
    fun notifyStateChanged(
        context: Context,
        providerClass: Class<*>,
        state: ConnectionState,
        buildViews: (ConnectionState) -> RemoteViews
    ) {
        val appWidgetManager = AppWidgetManager.getInstance(context) ?: return
        val appWidgetIds = appWidgetManager.getAppWidgetIds(ComponentName(context, providerClass))
        if (appWidgetIds.isEmpty()) return
        updateWidgets(appWidgetManager, appWidgetIds, state, buildViews)
    }
}
