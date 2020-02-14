package packet

type InventoryResult uint8

const (
	EQUIP_ERR_OK                               InventoryResult = iota
	EQUIP_ERR_CANT_EQUIP_LEVEL_I                               // You must reach level %d to use that item.
	EQUIP_ERR_CANT_EQUIP_SKILL                                 // You aren't skilled enough to use that item.
	EQUIP_ERR_WRONG_SLOT                                       // That item does not go in that slot.
	EQUIP_ERR_BAG_FULL                                         // That bag is full.
	EQUIP_ERR_BAG_IN_BAG                                       // Can't put non-empty bags in other bags.
	EQUIP_ERR_TRADE_EQUIPPED_BAG                               // You can't trade equipped bags.
	EQUIP_ERR_AMMO_ONLY                                        // Only ammo can go there.
	EQUIP_ERR_PROFICIENCY_NEEDED                               // You do not have the required proficiency for that item.
	EQUIP_ERR_NO_SLOT_AVAILABLE                                // No equipment slot is available for that item.
	EQUIP_ERR_CANT_EQUIP_EVER                                  // You can never use that item.
	EQUIP_ERR_CANT_EQUIP_EVER_2                                // You can never use that item.
	EQUIP_ERR_NO_SLOT_AVAILABLE_2                              // No equipment slot is available for that item.
	EQUIP_ERR_2HANDED_EQUIPPED                                 // Cannot equip that with a two-handed weapon.
	EQUIP_ERR_2HSKILLNOTFOUND                                  // You cannot dual-wield
	EQUIP_ERR_WRONG_BAG_TYPE                                   // That item doesn't go in that container.
	EQUIP_ERR_WRONG_BAG_TYPE_2                                 // That item doesn't go in that container.
	EQUIP_ERR_ITEM_MAX_COUNT                                   // You can't carry any more of those items.
	EQUIP_ERR_NO_SLOT_AVAILABLE_3                              // No equipment slot is available for that item.
	EQUIP_ERR_CANT_STACK                                       // This item cannot stack.
	EQUIP_ERR_NOT_EQUIPPABLE                                   // This item cannot be equipped.
	EQUIP_ERR_CANT_SWAP                                        // These items can't be swapped.
	EQUIP_ERR_SLOT_EMPTY                                       // That slot is empty.
	EQUIP_ERR_ITEM_NOT_FOUND                                   // The item was not found.
	EQUIP_ERR_DROP_BOUND_ITEM                                  // You can't drop a soulbound item.
	EQUIP_ERR_OUT_OF_RANGE                                     // Out of range.
	EQUIP_ERR_TOO_FEW_TO_SPLIT                                 // Tried to split more than number in stack.
	EQUIP_ERR_SPLIT_FAILED                                     // Couldn't split those items.
	EQUIP_ERR_SPELL_FAILED_REAGENTS_GENERIC                    // Missing reagent
	EQUIP_ERR_CANT_TRADE_GOLD                                  // Gold may only be offered by one trader.
	EQUIP_ERR_NOT_ENOUGH_MONEY                                 // You don't have enough money.
	EQUIP_ERR_NOT_A_BAG                                        // Not a bag.
	EQUIP_ERR_DESTROY_NONEMPTY_BAG                             // You can only do that with empty bags.
	EQUIP_ERR_NOT_OWNER                                        // You don't own that item.
	EQUIP_ERR_ONLY_ONE_QUIVER                                  // You can only equip one quiver.
	EQUIP_ERR_NO_BANK_SLOT                                     // You must purchase that bag slot first
	EQUIP_ERR_NO_BANK_HERE                                     // You are too far away from a bank.
	EQUIP_ERR_ITEM_LOCKED                                      // Item is locked.
	EQUIP_ERR_GENERIC_STUNNED                                  // You are stunned
	EQUIP_ERR_PLAYER_DEAD                                      // You can't do that when you're dead.
	EQUIP_ERR_CLIENT_LOCKED_OUT                                // You can't do that right now.
	EQUIP_ERR_INTERNAL_BAG_ERROR                               // Internal Bag Error
	EQUIP_ERR_ONLY_ONE_BOLT                                    // You can only equip one quiver.
	EQUIP_ERR_ONLY_ONE_AMMO                                    // You can only equip one ammo pouch.
	EQUIP_ERR_CANT_WRAP_STACKABLE                              // Stackable items can't be wrapped.
	EQUIP_ERR_CANT_WRAP_EQUIPPED                               // Equipped items can't be wrapped.
	EQUIP_ERR_CANT_WRAP_WRAPPED                                // Wrapped items can't be wrapped.
	EQUIP_ERR_CANT_WRAP_BOUND                                  // Bound items can't be wrapped.
	EQUIP_ERR_CANT_WRAP_UNIQUE                                 // Unique items can't be wrapped.
	EQUIP_ERR_CANT_WRAP_BAGS                                   // Bags can't be wrapped.
	EQUIP_ERR_LOOT_GONE                                        // Already looted
	EQUIP_ERR_INV_FULL                                         // Inventory is full.
	EQUIP_ERR_BANK_FULL                                        // Your bank is full
	EQUIP_ERR_VENDOR_SOLD_OUT                                  // That item is currently sold out.
	EQUIP_ERR_BAG_FULL_2                                       // That bag is full.
	EQUIP_ERR_ITEM_NOT_FOUND_2                                 // The item was not found.
	EQUIP_ERR_CANT_STACK_2                                     // This item cannot stack.
	EQUIP_ERR_BAG_FULL_3                                       // That bag is full.
	EQUIP_ERR_VENDOR_SOLD_OUT_2                                // That item is currently sold out.
	EQUIP_ERR_OBJECT_IS_BUSY                                   // That object is busy.
	EQUIP_ERR_CANT_BE_DISENCHANTED                             // Item cannot be disenchanted
	EQUIP_ERR_NOT_IN_COMBAT                                    // You can't do that while in combat
	EQUIP_ERR_NOT_WHILE_DISARMED                               // You can't do that while disarmed
	EQUIP_ERR_BAG_FULL_4                                       // That bag is full.
	EQUIP_ERR_CANT_EQUIP_RANK                                  // You don't have the required rank for that item
	EQUIP_ERR_CANT_EQUIP_REPUTATION                            // You don't have the required reputation for that item
	EQUIP_ERR_TOO_MANY_SPECIAL_BAGS                            // You cannot equip another bag of that type
	EQUIP_ERR_LOOT_CANT_LOOT_THAT_NOW                          // You can't loot that item now.
	EQUIP_ERR_ITEM_UNIQUE_EQUIPPABLE                           // You cannot equip more than one of those.
	EQUIP_ERR_VENDOR_MISSING_TURNINS                           // You do not have the required items for that purchase
	EQUIP_ERR_NOT_ENOUGH_HONOR_POINTS                          // You don't have enough honor points
	EQUIP_ERR_NOT_ENOUGH_ARENA_POINTS                          // You don't have enough arena points
	EQUIP_ERR_ITEM_MAX_COUNT_SOCKETED                          // You have the maximum number of those gems in your inventory or socketed into items.
	EQUIP_ERR_MAIL_BOUND_ITEM                                  // You can't mail soulbound items.
	EQUIP_ERR_INTERNAL_BAG_ERROR_2                             // Internal Bag Error
	EQUIP_ERR_BAG_FULL_5                                       // That bag is full.
	EQUIP_ERR_ITEM_MAX_COUNT_EQUIPPED_SOCKETED                 // You have the maximum number of those gems socketed into equipped items.
	EQUIP_ERR_ITEM_UNIQUE_EQUIPPABLE_SOCKETED                  // You cannot socket more than one of those gems into a single item.
	EQUIP_ERR_TOO_MUCH_GOLD                                    // At gold limit
	EQUIP_ERR_NOT_DURING_ARENA_MATCH                           // You can't do that while in an arena match
	EQUIP_ERR_TRADE_BOUND_ITEM                                 // You can't trade a soulbound item.
	EQUIP_ERR_CANT_EQUIP_RATING                                // You don't have the personal, team, or battleground rating required to buy that item
	EQUIP_ERR_EVENT_AUTOEQUIP_BIND_CONFIRM
	EQUIP_ERR_NOT_SAME_ACCOUNT // Account-bound items can only be given to your own characters.
	EQUIP_NONE_3
	EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_COUNT_EXCEEDED_IS    // You can only carry %d %s
	EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_SOCKETED_EXCEEDED_IS // You can only equip %d |4item:items in the %s category
	EQUIP_ERR_SCALING_STAT_ITEM_LEVEL_EXCEEDED             // Your level is too high to use that item
	EQUIP_ERR_PURCHASE_LEVEL_TOO_LOW                       // You must reach level %d to purchase that item.
	EQUIP_ERR_CANT_EQUIP_NEED_TALENT                       // You do not have the required talent to equip that.
	EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_EQUIPPED_EXCEEDED_IS // You can only equip %d |4item:items in the %s category
	EQUIP_ERR_SHAPESHIFT_FORM_CANNOT_EQUIP                 // Cannot equip item in this form
	EQUIP_ERR_ITEM_INVENTORY_FULL_SATCHEL                  // Your inventory is full. Your satchel has been delivered to your mailbox.
	EQUIP_ERR_SCALING_STAT_ITEM_LEVEL_TOO_LOW              // Your level is too low to use that item
	EQUIP_ERR_CANT_BUY_QUANTITY                            // You can't buy the specified quantity of that item.
	EQUIP_ERR_ITEM_IS_BATTLE_PAY_LOCKED                    // Your purchased item is still waiting to be unlocked
	EQUIP_ERR_REAGENT_BANK_FULL                            // Your reagent bank is full
	EQUIP_ERR_REAGENT_BANK_LOCKED
	EQUIP_ERR_WRONG_BAG_TYPE_3         // That item doesn't go in that container.
	EQUIP_ERR_CANT_USE_ITEM            // You can't use that item.
	EQUIP_ERR_CANT_BE_OBLITERATED      // You can't obliterate that item
	EQUIP_ERR_GUILD_BANK_CONJURED_ITEM // You cannot store conjured items in the guild bank
	EQUIP_ERR_CANT_DO_THAT_RIGHT_NOW   // You can't do that right now.
	EQUIP_ERR_BAG_FULL_6               // That bag is full.
	EQUIP_ERR_CANT_BE_SCRAPPED         // You can't scrap that item
	EQUIP_NONE_4
)

type InventoryResultDescriptor map[InventoryResult]uint8

var InventoryResultDescriptors = map[uint32]InventoryResultDescriptor{
	5875: {
		EQUIP_ERR_OK:                            0,
		EQUIP_ERR_CANT_EQUIP_LEVEL_I:            1,
		EQUIP_ERR_CANT_EQUIP_SKILL:              2,
		EQUIP_ERR_WRONG_SLOT:                    3,
		EQUIP_ERR_BAG_FULL:                      4,
		EQUIP_ERR_BAG_IN_BAG:                    5,
		EQUIP_ERR_TRADE_EQUIPPED_BAG:            6,
		EQUIP_ERR_AMMO_ONLY:                     7,
		EQUIP_ERR_PROFICIENCY_NEEDED:            8,
		EQUIP_ERR_NO_SLOT_AVAILABLE:             9,
		EQUIP_ERR_CANT_EQUIP_EVER:               10,
		EQUIP_ERR_CANT_EQUIP_EVER_2:             11,
		EQUIP_ERR_NO_SLOT_AVAILABLE_2:           12,
		EQUIP_ERR_2HANDED_EQUIPPED:              13,
		EQUIP_ERR_2HSKILLNOTFOUND:               14,
		EQUIP_ERR_WRONG_BAG_TYPE:                15,
		EQUIP_ERR_WRONG_BAG_TYPE_2:              16,
		EQUIP_ERR_ITEM_MAX_COUNT:                17,
		EQUIP_ERR_NO_SLOT_AVAILABLE_3:           18,
		EQUIP_ERR_CANT_STACK:                    19,
		EQUIP_ERR_NOT_EQUIPPABLE:                20,
		EQUIP_ERR_CANT_SWAP:                     21,
		EQUIP_ERR_SLOT_EMPTY:                    22,
		EQUIP_ERR_ITEM_NOT_FOUND:                23,
		EQUIP_ERR_DROP_BOUND_ITEM:               24,
		EQUIP_ERR_OUT_OF_RANGE:                  25,
		EQUIP_ERR_TOO_FEW_TO_SPLIT:              26,
		EQUIP_ERR_SPLIT_FAILED:                  27,
		EQUIP_ERR_SPELL_FAILED_REAGENTS_GENERIC: 28,
		EQUIP_ERR_NOT_ENOUGH_MONEY:              29,
		EQUIP_ERR_NOT_A_BAG:                     30,
		EQUIP_ERR_DESTROY_NONEMPTY_BAG:          31,
		EQUIP_ERR_NOT_OWNER:                     32,
		EQUIP_ERR_ONLY_ONE_QUIVER:               33,
		EQUIP_ERR_NO_BANK_SLOT:                  34,
		EQUIP_ERR_NO_BANK_HERE:                  35,
		EQUIP_ERR_ITEM_LOCKED:                   36,
		EQUIP_ERR_GENERIC_STUNNED:               37,
		EQUIP_ERR_PLAYER_DEAD:                   38,
		EQUIP_ERR_CANT_DO_THAT_RIGHT_NOW:        39,
		EQUIP_ERR_INTERNAL_BAG_ERROR:            40,
		EQUIP_ERR_ONLY_ONE_BOLT:                 41,
		EQUIP_ERR_ONLY_ONE_AMMO:                 42,
		EQUIP_ERR_CANT_WRAP_STACKABLE:           43,
		EQUIP_ERR_CANT_WRAP_EQUIPPED:            44,
		EQUIP_ERR_CANT_WRAP_WRAPPED:             45,
		EQUIP_ERR_CANT_WRAP_BOUND:               46,
		EQUIP_ERR_CANT_WRAP_UNIQUE:              47,
		EQUIP_ERR_CANT_WRAP_BAGS:                48,
		EQUIP_ERR_LOOT_GONE:                     49,
		EQUIP_ERR_INV_FULL:                      50,
		EQUIP_ERR_BANK_FULL:                     51,
		EQUIP_ERR_VENDOR_SOLD_OUT:               52,
		EQUIP_ERR_BAG_FULL_2:                    53,
		EQUIP_ERR_ITEM_NOT_FOUND_2:              54,
		EQUIP_ERR_CANT_STACK_2:                  55,
		EQUIP_ERR_BAG_FULL_3:                    56,
		EQUIP_ERR_VENDOR_SOLD_OUT_2:             57,
		EQUIP_ERR_OBJECT_IS_BUSY:                58,
		EQUIP_NONE_3:                            59,
		EQUIP_ERR_NOT_IN_COMBAT:                 60,
		EQUIP_ERR_NOT_WHILE_DISARMED:            61,
		EQUIP_ERR_BAG_FULL_4:                    62,
		EQUIP_ERR_CANT_EQUIP_RANK:               63,
		EQUIP_ERR_CANT_EQUIP_REPUTATION:         64,
		EQUIP_ERR_TOO_MANY_SPECIAL_BAGS:         65,
		EQUIP_ERR_LOOT_CANT_LOOT_THAT_NOW:       66,
	},

	32978: {
		EQUIP_ERR_OK:                                           0,
		EQUIP_ERR_CANT_EQUIP_LEVEL_I:                           1,
		EQUIP_ERR_CANT_EQUIP_SKILL:                             2,
		EQUIP_ERR_WRONG_SLOT:                                   3,
		EQUIP_ERR_BAG_FULL:                                     4,
		EQUIP_ERR_BAG_IN_BAG:                                   5,
		EQUIP_ERR_TRADE_EQUIPPED_BAG:                           6,
		EQUIP_ERR_AMMO_ONLY:                                    7,
		EQUIP_ERR_PROFICIENCY_NEEDED:                           8,
		EQUIP_ERR_NO_SLOT_AVAILABLE:                            9,
		EQUIP_ERR_CANT_EQUIP_EVER:                              10,
		EQUIP_ERR_CANT_EQUIP_EVER_2:                            11,
		EQUIP_ERR_NO_SLOT_AVAILABLE_2:                          12,
		EQUIP_ERR_2HANDED_EQUIPPED:                             13,
		EQUIP_ERR_2HSKILLNOTFOUND:                              14,
		EQUIP_ERR_WRONG_BAG_TYPE:                               15,
		EQUIP_ERR_WRONG_BAG_TYPE_2:                             16,
		EQUIP_ERR_ITEM_MAX_COUNT:                               17,
		EQUIP_ERR_NO_SLOT_AVAILABLE_3:                          18,
		EQUIP_ERR_CANT_STACK:                                   19,
		EQUIP_ERR_NOT_EQUIPPABLE:                               20,
		EQUIP_ERR_CANT_SWAP:                                    21,
		EQUIP_ERR_SLOT_EMPTY:                                   22,
		EQUIP_ERR_ITEM_NOT_FOUND:                               23,
		EQUIP_ERR_DROP_BOUND_ITEM:                              24,
		EQUIP_ERR_OUT_OF_RANGE:                                 25,
		EQUIP_ERR_TOO_FEW_TO_SPLIT:                             26,
		EQUIP_ERR_SPLIT_FAILED:                                 27,
		EQUIP_ERR_SPELL_FAILED_REAGENTS_GENERIC:                28,
		EQUIP_ERR_CANT_TRADE_GOLD:                              29,
		EQUIP_ERR_NOT_ENOUGH_MONEY:                             30,
		EQUIP_ERR_NOT_A_BAG:                                    31,
		EQUIP_ERR_DESTROY_NONEMPTY_BAG:                         32,
		EQUIP_ERR_NOT_OWNER:                                    33,
		EQUIP_ERR_ONLY_ONE_QUIVER:                              34,
		EQUIP_ERR_NO_BANK_SLOT:                                 35,
		EQUIP_ERR_NO_BANK_HERE:                                 36,
		EQUIP_ERR_ITEM_LOCKED:                                  37,
		EQUIP_ERR_GENERIC_STUNNED:                              38,
		EQUIP_ERR_PLAYER_DEAD:                                  39,
		EQUIP_ERR_CLIENT_LOCKED_OUT:                            40,
		EQUIP_ERR_INTERNAL_BAG_ERROR:                           41,
		EQUIP_ERR_ONLY_ONE_BOLT:                                42,
		EQUIP_ERR_ONLY_ONE_AMMO:                                43,
		EQUIP_ERR_CANT_WRAP_STACKABLE:                          44,
		EQUIP_ERR_CANT_WRAP_EQUIPPED:                           45,
		EQUIP_ERR_CANT_WRAP_WRAPPED:                            46,
		EQUIP_ERR_CANT_WRAP_BOUND:                              47,
		EQUIP_ERR_CANT_WRAP_UNIQUE:                             48,
		EQUIP_ERR_CANT_WRAP_BAGS:                               49,
		EQUIP_ERR_LOOT_GONE:                                    50,
		EQUIP_ERR_INV_FULL:                                     51,
		EQUIP_ERR_BANK_FULL:                                    52,
		EQUIP_ERR_VENDOR_SOLD_OUT:                              53,
		EQUIP_ERR_BAG_FULL_2:                                   54,
		EQUIP_ERR_ITEM_NOT_FOUND_2:                             55,
		EQUIP_ERR_CANT_STACK_2:                                 56,
		EQUIP_ERR_BAG_FULL_3:                                   57,
		EQUIP_ERR_VENDOR_SOLD_OUT_2:                            58,
		EQUIP_ERR_OBJECT_IS_BUSY:                               59,
		EQUIP_ERR_CANT_BE_DISENCHANTED:                         60,
		EQUIP_ERR_NOT_IN_COMBAT:                                61,
		EQUIP_ERR_NOT_WHILE_DISARMED:                           62,
		EQUIP_ERR_BAG_FULL_4:                                   63,
		EQUIP_ERR_CANT_EQUIP_RANK:                              64,
		EQUIP_ERR_CANT_EQUIP_REPUTATION:                        65,
		EQUIP_ERR_TOO_MANY_SPECIAL_BAGS:                        66,
		EQUIP_ERR_LOOT_CANT_LOOT_THAT_NOW:                      67,
		EQUIP_ERR_ITEM_UNIQUE_EQUIPPABLE:                       68,
		EQUIP_ERR_VENDOR_MISSING_TURNINS:                       69,
		EQUIP_ERR_NOT_ENOUGH_HONOR_POINTS:                      70,
		EQUIP_ERR_NOT_ENOUGH_ARENA_POINTS:                      71,
		EQUIP_ERR_ITEM_MAX_COUNT_SOCKETED:                      72,
		EQUIP_ERR_MAIL_BOUND_ITEM:                              73,
		EQUIP_ERR_INTERNAL_BAG_ERROR_2:                         74,
		EQUIP_ERR_BAG_FULL_5:                                   75,
		EQUIP_ERR_ITEM_MAX_COUNT_EQUIPPED_SOCKETED:             76,
		EQUIP_ERR_ITEM_UNIQUE_EQUIPPABLE_SOCKETED:              77,
		EQUIP_ERR_TOO_MUCH_GOLD:                                78,
		EQUIP_ERR_NOT_DURING_ARENA_MATCH:                       79,
		EQUIP_ERR_TRADE_BOUND_ITEM:                             80,
		EQUIP_ERR_CANT_EQUIP_RATING:                            81,
		EQUIP_ERR_EVENT_AUTOEQUIP_BIND_CONFIRM:                 82,
		EQUIP_ERR_NOT_SAME_ACCOUNT:                             83,
		EQUIP_NONE_3:                                           84,
		EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_COUNT_EXCEEDED_IS:    85,
		EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_SOCKETED_EXCEEDED_IS: 86,
		EQUIP_ERR_SCALING_STAT_ITEM_LEVEL_EXCEEDED:             87,
		EQUIP_ERR_PURCHASE_LEVEL_TOO_LOW:                       88,
		EQUIP_ERR_CANT_EQUIP_NEED_TALENT:                       89,
		EQUIP_ERR_ITEM_MAX_LIMIT_CATEGORY_EQUIPPED_EXCEEDED_IS: 90,
		EQUIP_ERR_SHAPESHIFT_FORM_CANNOT_EQUIP:                 91,
		EQUIP_ERR_ITEM_INVENTORY_FULL_SATCHEL:                  92,
		EQUIP_ERR_SCALING_STAT_ITEM_LEVEL_TOO_LOW:              93,
		EQUIP_ERR_CANT_BUY_QUANTITY:                            94,
		EQUIP_ERR_ITEM_IS_BATTLE_PAY_LOCKED:                    95,
		EQUIP_ERR_REAGENT_BANK_FULL:                            96,
		EQUIP_ERR_REAGENT_BANK_LOCKED:                          97,
		EQUIP_ERR_WRONG_BAG_TYPE_3:                             98,
		EQUIP_ERR_CANT_USE_ITEM:                                99,
		EQUIP_ERR_CANT_BE_OBLITERATED:                          100, // You can't obliterate that item
		EQUIP_ERR_GUILD_BANK_CONJURED_ITEM:                     101, // You cannot store conjured items in the guild bank
		EQUIP_ERR_CANT_DO_THAT_RIGHT_NOW:                       102, // You can't do that right now.
		EQUIP_ERR_BAG_FULL_6:                                   103, // That bag is full.
		EQUIP_ERR_CANT_BE_SCRAPPED:                             104, // You can't scrap that item
		EQUIP_NONE_4:                                           105,
	},
}
