// package externaltools contains a group of tools outside go runtime.
// They are provided by the runtime environment. Some of them must be initiazlied before using, and be cleaned up properly when finishing.
// The dependency of packages under externaltools must be as less as possible, to avoid cycle dependencies.
package externaltools
