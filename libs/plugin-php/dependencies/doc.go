// Package dependencies lists and changes the Composer packages a PHP
// workspace declares: the require and require-dev of each package's
// composer.json and the workspace root's, so `smith sync`, `smith up` and
// the one-version check cover PHP. Installing the workspace and adding or
// removing a requirement go through Composer.
package dependencies
