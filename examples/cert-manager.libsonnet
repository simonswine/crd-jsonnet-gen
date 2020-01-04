{
  "meta.cert-manager.io":: {
    local apiGroup = 'meta.cert-manager.io',
    v1:: {
      local apiVersion = {
        apiVersion: '%s/v1' % apiGroup,
      },
    },
  },
  "cert-manager.io":: {
    local apiGroup = 'cert-manager.io',
    v1alpha2:: {
      local apiVersion = {
        apiVersion: '%s/v1alpha2' % apiGroup,
      },
      certificate:: {
        local kind = {
          kind: 'Certificate',
        },
        new():: kind + apiVersion,
      },
      certificateRequest:: {
        local kind = {
          kind: 'CertificateRequest',
        },
        new():: kind + apiVersion,
      },
      issuer:: {
        local kind = {
          kind: 'Issuer',
        },
        new():: kind + apiVersion,
      },
      clusterIssuer:: {
        local kind = {
          kind: 'ClusterIssuer',
        },
        new():: kind + apiVersion,
      },
    },
  },
  "acme.cert-manager.io":: {
    local apiGroup = 'acme.cert-manager.io',
    v1alpha2:: {
      local apiVersion = {
        apiVersion: '%s/v1alpha2' % apiGroup,
      },
      challenge:: {
        local kind = {
          kind: 'Challenge',
        },
        new():: kind + apiVersion,
      },
      order:: {
        local kind = {
          kind: 'Order',
        },
        new():: kind + apiVersion,
      },
    },
  },
}