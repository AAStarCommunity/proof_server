package base1024

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeString(t *testing.T) {
	t.Run("Equally", func(t *testing.T) {
		str := "🐟🔂🏁🤖💧🚊😤"
		res, err := DecodeString(str)
		assert.Nil(t, err)
		assert.Equal(t, "Maskbook", string(res))
	})

	t.Run("Not Equal", func(t *testing.T) {
		str := "🐟🔂🏁🤖💧"
		res, err := DecodeString(str)
		fmt.Println(string(res), err)
		assert.Nil(t, err)
		assert.NotEqual(t, "Maskbook", string(res))
	})

	t.Run("Decode Playload", func(t *testing.T) {
		str := "👲🍚🍾🔆🏠🚱👧🦢🕟🛷🔭💘😝🙂🚳🦜🔙🔊🚗🏏👪🛹🗣🏳🦐🥫🦺🚎🕗🚷💡🚁🎟🗯📰🐊🕳🥠💐🎛🤵🆘🔣📥🦝🔉🌊🥠🥅🍏🥜🃏"
		//"👲🍚🍾🔆🏠🚱👧\U0001F9A2🕟🛷🔭💘😝🙂🚳\U0001F99C🔙🔊🚗🏏👪\U0001F6F9🗣🏳🦐🥫\U0001F9BA🚎🕗🚷💡🚁🎟🗯📰🐊🕳🥠💐🎛🤵🆘🔣📥\U0001F99D🔉🌊🥠🥅🍏🥜🃏"
		//str := "👲🍚🍾🔆🏠🚱👧\U0001F9A2🕟🛷🔭💘😝🙂🚳\U0001F99C🔙🔊🚗🏏👪\U0001F6F9🗣🏳🦐🥫\U0001F9BA🚎🕗🚷💡🚁🎟🗯📰🐊🕳🥠💐🎛🤵🆘🔣📥\U0001F99D🔉🌊🥠🥅🍏🥜🃏"
		res, err := DecodeString(str)
		fmt.Println(string(res), err)
		assert.Nil(t, err)
		assert.NotEqual(t, "Maskbook", string(res))
	})
}
