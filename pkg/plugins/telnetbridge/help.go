package telnetbridge

import "strings"

const helpText = "Mqtt 2 Telnet Bridge: \n" +
	" *** list -> show all active connections \n" +
	" *** connect {ip} {port} -> open telnet connection \n" +
	" *** disconnect -> close telnet connection"

const errorText = "***: command not valid, try > *** help"

func getHelpText(pluginName string) string {
	return strings.Replace(helpText, "***", pluginName, -1)
}

func getErrorText(pluginName string) string {
	return strings.Replace(errorText, "***", pluginName, -1)
}
