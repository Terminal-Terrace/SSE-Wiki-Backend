import type { RpcClientConfig } from '@sse-wiki/rpc-client'
import path from 'node:path'
import { RpcClient } from '@sse-wiki/rpc-client'
import type * as AuthServiceTypes from './types/auth_service'

/**
 * 服务客户端配置
 */
export interface ServiceConfig {
  serverAddress?: string
  /** 其他 RpcClient 配置 */
  rpcConfig?: Partial<Omit<RpcClientConfig, 'protoPath' | 'packageName' | 'serviceClassName' | 'serverAddress'>>
}

/**
 * AuthService 客户端
 *
 * 使用示例：
 * ```typescript
 * const client = new AuthService()
 * const response = await client.GetUser({ id: 123 })
 * ```
 */
export class AuthService {
  private rpcClient: RpcClient

  /**
   * 创建 AuthService 客户端
   * @param config 可选配置
   */
  constructor(config?: ServiceConfig) {
    const protoPath = path.join(import.meta.dirname, 'proto', 'auth_service', 'auth_service.proto')

    this.rpcClient = new RpcClient({
      protoPath,
      packageName: 'auth_service',
      serviceClassName: 'AuthService',
      serverAddress: config?.serverAddress || 'localhost:50051',
      ...config?.rpcConfig,
    })
  }

  async Prelogin(req: AuthServiceTypes.PreloginRequest): Promise<AuthServiceTypes.PreloginResponse> {
    return this.rpcClient.call<AuthServiceTypes.PreloginRequest, AuthServiceTypes.PreloginResponse>('Prelogin', req)
  }

  async SendCode(req: AuthServiceTypes.CodeRequest): Promise<AuthServiceTypes.CodeResponse> {
    return this.rpcClient.call<AuthServiceTypes.CodeRequest, AuthServiceTypes.CodeResponse>('SendCode', req)
  }

  async Login(req: AuthServiceTypes.LoginRequest): Promise<AuthServiceTypes.LoginResponse> {
    return this.rpcClient.call<AuthServiceTypes.LoginRequest, AuthServiceTypes.LoginResponse>('Login', req)
  }

  async Logout(req: AuthServiceTypes.LogoutRequest): Promise<AuthServiceTypes.LogoutResponse> {
    return this.rpcClient.call<AuthServiceTypes.LogoutRequest, AuthServiceTypes.LogoutResponse>('Logout', req)
  }

  async GetUserInfo(req: AuthServiceTypes.InfoRequest): Promise<AuthServiceTypes.InfoResponse> {
    return this.rpcClient.call<AuthServiceTypes.InfoRequest, AuthServiceTypes.InfoResponse>('GetUserInfo', req)
  }

  async RefreshToken(req: AuthServiceTypes.RefreshRequest): Promise<AuthServiceTypes.RefreshResponse> {
    return this.rpcClient.call<AuthServiceTypes.RefreshRequest, AuthServiceTypes.RefreshResponse>('RefreshToken', req)
  }

  async Register(req: AuthServiceTypes.RegisterRequest): Promise<AuthServiceTypes.RegisterResponse> {
    return this.rpcClient.call<AuthServiceTypes.RegisterRequest, AuthServiceTypes.RegisterResponse>('Register', req)
  }

  async UpdateProfile(req: AuthServiceTypes.UpdateProfileRequest): Promise<AuthServiceTypes.UpdateProfileResponse> {
    return this.rpcClient.call<AuthServiceTypes.UpdateProfileRequest, AuthServiceTypes.UpdateProfileResponse>('UpdateProfile', req)
  }

  /**
   * 关闭 gRPC 连接
   */
  close(): void {
    this.rpcClient.close()
  }

  /**
   * 等待连接就绪
   * @param deadline 超时时间（毫秒）
   */
  async waitForReady(deadline?: number): Promise<void> {
    return this.rpcClient.waitForReady(deadline)
  }
}
