package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"fss-mining/internal/model"
	"fss-mining/internal/repository"
	pkgjwt "fss-mining/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db          *gorm.DB
	userRepo    *repository.UserRepo
	accountRepo *repository.AccountRepo
}

func NewUserService(db *gorm.DB, ur *repository.UserRepo, ar *repository.AccountRepo) *UserService {
	return &UserService{db: db, userRepo: ur, accountRepo: ar}
}

type RegisterReq struct {
	Username   string
	Password   string
	Email      string
	Phone      string
	InviteCode string // 邀请码（填写邀请人的邀请码）
}

type LoginResp struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *UserService) Register(ctx context.Context, req RegisterReq) (*model.User, error) {
	// 检查用户名是否已存在
	existUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查邮箱
	if req.Email != "" {
		existEmail, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if existEmail != nil {
			return nil, fmt.Errorf("邮箱已注册")
		}
	}

	// 哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 生成邀请码
	inviteCode, err := genInviteCode()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:   req.Username,
		Password:   string(hash),
		Email:      req.Email,
		Phone:      req.Phone,
		Role:       model.UserRoleUser,
		Status:     model.UserStatusActive,
		InviteCode: inviteCode,
	}

	// 查找邀请人
	if req.InviteCode != "" {
		referrer, err := s.userRepo.GetByInviteCode(ctx, req.InviteCode)
		if err != nil {
			return nil, err
		}
		if referrer != nil {
			user.ReferrerID = &referrer.ID
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}
		// 自动创建交易区和挖矿区账户
		if _, err := s.accountRepo.GetOrCreate(ctx, user.ID, model.AccountTypeTrade); err != nil {
			return err
		}
		if _, err := s.accountRepo.GetOrCreate(ctx, user.ID, model.AccountTypeMining); err != nil {
			return err
		}
		return nil
	})

	return user, err
}

func (s *UserService) Login(ctx context.Context, username, password string) (*LoginResp, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("用户名或密码错误")
	}
	if user.Status != model.UserStatusActive {
		return nil, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := pkgjwt.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResp{Token: token, User: user}, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func genInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
