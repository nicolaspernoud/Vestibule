package tokens

import (
	"fmt"
	"testing"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/common"
)

type user struct {
	Login    string
	Password string
}

func (u user) String() string {
	return fmt.Sprintf("Login: %v, Password: %v", u.Login, u.Password)
}

func Test_manager_CreateToken_unStoreData(t *testing.T) {
	key, _ := common.GenerateRandomBytes(32)
	key2, _ := common.GenerateRandomBytes(32)
	type fields struct {
		encryptKey []byte
		decryptKey []byte
		debugMode  bool
	}
	type args struct {
		data       interface{}
		expiration time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{"future_expiration", fields{key, key, false}, args{user{"admin", "password"}, time.Now().Add(24 * time.Hour)}, true, false},
		{"past_expiration", fields{key, key, false}, args{user{"admin", "password"}, time.Now().Add(-24 * time.Hour)}, false, true},
		{"incorrect_aes_key", fields{[]byte("wrong_key_size"), []byte("wrong_key_size"), false}, args{user{"admin", "password"}, time.Now().Add(+24 * time.Hour)}, false, true},
		{"wrong_decrypt_key", fields{key, key2, false}, args{user{"admin", "password"}, time.Now().Add(+24 * time.Hour)}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := manager{
				key:       tt.fields.encryptKey,
				debugMode: tt.fields.debugMode,
			}
			token, _ := m.CreateToken(tt.args.data, tt.args.expiration)
			m.key = tt.fields.decryptKey
			v := user{}
			err := m.unstoreData(token, &v)
			got := tt.args.data == v
			if (err != nil) != tt.wantErr {
				t.Errorf("manager.(un)storeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("manager.(un)storeData() inData:%v, outData:%v => equality: %v, want %v", tt.args.data, v, got, tt.want)
			}
		})
	}
}
