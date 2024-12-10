package model

import (
	"testing"
)

func TestNacosHttpSdk_UpdateServiceHealthCheckTypeToNull(t *testing.T) {
	type fields struct {
		nacosAddresses []string
		nacosNamespace string
		accessKey      string
		secretKey      string
	}
	type args struct {
		key ServiceKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				nacosAddresses: []string{"127.0.0.1:8848"},
				nacosNamespace: "public",
				accessKey:      "",
				secretKey:      "",
			},
			args: args{
				key: ServiceKey{
					ServiceName: "test",
					Group:       "DEFAULT_GROUP",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NacosHttpSdk{
				nacosAddresses: tt.fields.nacosAddresses,
				nacosNamespace: tt.fields.nacosNamespace,
				accessKey:      tt.fields.accessKey,
				secretKey:      tt.fields.secretKey,
			}
			if got := n.UpdateServiceHealthCheckTypeToNone(tt.args.key); got != tt.want {
				t.Errorf("UpdateServiceHealthCheckTypeToNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNacosHttpSdk_getService(t *testing.T) {
	type fields struct {
		nacosAddresses []string
		nacosNamespace string
		accessKey      string
		secretKey      string
	}
	type args struct {
		serviceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ServiceDetail
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				nacosAddresses: []string{"127.0.0.1:8848"},
				nacosNamespace: "public",
				accessKey:      "",
				secretKey:      "",
			},
			args: args{
				serviceName: "test",
			},
			want: ServiceDetail{
				GroupName: "DEFAUTL_GROUP",
				Name:      "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NacosHttpSdk{
				nacosAddresses: tt.fields.nacosAddresses,
				nacosNamespace: tt.fields.nacosNamespace,
				accessKey:      tt.fields.accessKey,
				secretKey:      tt.fields.secretKey,
			}
			_, err := n.getService(tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
