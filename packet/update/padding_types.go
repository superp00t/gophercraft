package update

// These padding types are not for handling unknown types.
// Padding types are for handling empty fields, or fields used to align data with the chunk stream.
// If your field's purpose is unknown, give it an Unk(x) name.
// This is bad practice though. Often server projects for future game versions know the true names for these fields, and can be added to your descriptor.

// ChunkPad will move the chunk offset 1 forward. If the decoder detects an enabled value at a ChunkPad's offset, it will return an error.
type ChunkPad struct{}

// BitPad will move the bit offset 1 forward. If a detected value is 1 at this location, it will not return an error.
type BitPad struct{}

// BytePad will move the bit offset 1 byte (8 bits) forward. In other words, if your bit offset is not aligned with a byte stream, it will move toward the next offset divisible by 8.
type BytePad struct{}
