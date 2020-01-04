{
  "contour.heptio.com":: {
    local apiGroup = 'contour.heptio.com',
    v1beta1:: {
      local apiVersion = {
        apiVersion: '%s/v1beta1' % apiGroup,
      },
      ingressRoute:: {
        local kind = {
          kind: 'IngressRoute',
        },
        new():: kind + apiVersion,
      },
      tLSCertificateDelegation:: {
        local kind = {
          kind: 'TLSCertificateDelegation',
        },
        new():: kind + apiVersion,
      },
    },
  },
}