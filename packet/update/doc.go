//Package update provides functions for serializing and deserializing SMSG_UPDATE_OBJECT.
//SMSG_UPDATE_OBJECT notifies the game client of incremental state changes, or updates, to in-world objects.
//The structure of this packet varies extremely across protocol revisions, so this package incorporates several descriptor modules for storing fields.
package update
